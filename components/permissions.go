/*
Copyright 2020 Noah Kantrowitz
Copyright 2019-2020 Ridecell, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package components

import (
	"context"

	cu "github.com/coderanger/controller-utils"
	"github.com/go-logr/logr"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

type permissionsComponent struct {
	clientFactory rabbitClientFactory
}

type permissionsComponentWatchMap struct {
	client client.Client
	log    logr.Logger
}

func Permissions() *permissionsComponent {
	return &permissionsComponent{clientFactory: rabbitholeClientFactory}
}

func (comp *permissionsComponent) Setup(ctx *cu.Context, bldr *ctrl.Builder) error {
	bldr.Watches(
		&source.Kind{Type: &rabbitv1beta1.RabbitVhost{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: &permissionsComponentWatchMap{client: ctx.Client, log: ctx.Log}},
	)
	return nil
}

// Watch map function used above.
// Obj is a Vhost that just got an event, map it back to any User with * permissions.
func (wm *permissionsComponentWatchMap) Map(obj handler.MapObject) []reconcile.Request {
	requests := []reconcile.Request{}
	// Find any User objects that have * vhost permissions so they can be updated.
	users := &rabbitv1beta1.RabbitUserList{}
	err := wm.client.List(context.Background(), users)
	if err != nil {
		wm.log.Error(err, "error listing users")
		// TODO Metric to track this for alerting.
		return requests
	}
	for _, user := range users.Items {
		for _, perm := range user.Spec.Permissions {
			if perm.Vhost == "*" {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      user.Name,
						Namespace: user.Namespace,
					},
				})
				break
			}
		}
	}
	return requests
}

func (comp *permissionsComponent) Reconcile(ctx *cu.Context) (cu.Result, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitUser)
	ctx.Conditions.SetUnknown("PermissionsReady", "Unknown")

	// Connect to the RabbitMQ server.
	rmqc, _, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	// Get the core data for the user from the object/context.
	username := obj.Spec.Username
	if username == "" { // TODO Switch this to a defaulting webhook.
		username = obj.Name
	}

	// Look for a `*` vhost in the spec, move the rest into a holding pen.
	specPermMap := map[string]*rabbitv1beta1.RabbitPermission{}
	var allVhostPerm *rabbitv1beta1.RabbitPermission
	for _, perm := range obj.Spec.Permissions {
		permCopy := perm
		if perm.Vhost == "*" {
			allVhostPerm = &permCopy
		} else {
			specPermMap[perm.Vhost] = &permCopy
		}
	}
	if allVhostPerm != nil {
		// Expand the * pseudo-vhost.
		vhosts, err := rmqc.ListVhosts()
		if err != nil {
			return cu.Result{}, errors.Wrap(err, "error listing vhosts for * vhost permissions")
		}
		for _, vhost := range vhosts {
			_, alreadySet := specPermMap[vhost.Name]
			if !alreadySet {
				specPermMap[vhost.Name] = allVhostPerm
			}
		}
	}

	// Get all Permissions for a vhost, user. Add all mentioned in spec and remove unwanted.
	permissions, err := rmqc.ListPermissionsOf(username)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error listing permissions for user %s", username)
	}
	existingPermMap := map[string]*rabbithole.PermissionInfo{}
	for _, perm := range permissions {
		existingPermMap[perm.Vhost] = &perm
	}

	for vhostName, perm := range specPermMap {
		var createPermissions, updatePermissions bool

		existingPerm, ok := existingPermMap[vhostName]
		if !ok {
			createPermissions = true
		} else {
			// Delete the entry in permMap so we can use it as a double-ended diff too.
			delete(existingPermMap, vhostName)
			updatePermissions = existingPerm.Read != perm.Read || existingPerm.Write != perm.Write || existingPerm.Configure != perm.Configure
		}

		if createPermissions || updatePermissions {
			_, err := rmqc.UpdatePermissionsIn(vhostName, username, rabbithole.Permissions{
				Configure: perm.Configure,
				Read:      perm.Read,
				Write:     perm.Write,
			})
			if err != nil {
				return cu.Result{}, errors.Wrapf(err, "error updating permissions for user %s and vhost %s", username, vhostName)
			}

			// Create an event.
			var event, eventMessage string
			if createPermissions {
				event = "PermissionsCreated"
				eventMessage = "created"
			} else {
				event = "PermissionsUpdated"
				eventMessage = "updated"
			}
			ctx.Events.Eventf(obj, "Normal", event, "RabbitMQ permissions for user %s in vhost %s %s", username, vhostName, eventMessage)
		}
	}

	//Remove any permissions that exist in RabbitMQ but not in the Spec.
	for vhost := range existingPermMap {
		// 204 response code when permission is removed.
		_, err := rmqc.ClearPermissionsIn(vhost, username)
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error removing permissions for user %s and vhost %s", username, vhost)
		}
		ctx.Events.Eventf(obj, "Normal", "PermissionsDeleted", "RabbitMQ permissions for user %s in vhost %s deleted", username, vhost)

	}

	ctx.Conditions.SetTrue("PermissionsReady", "PermissionsSynced")
	return cu.Result{}, nil
}
