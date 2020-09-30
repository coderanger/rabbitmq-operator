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
	cu "github.com/coderanger/controller-utils"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	"github.com/pkg/errors"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

type permissionsComponent struct {
	clientFactory rabbitClientFactory
}

func Permissions() *permissionsComponent {
	return &permissionsComponent{clientFactory: rabbitholeClientFactory}
}

func (comp *permissionsComponent) Reconcile(ctx *cu.Context) (cu.Result, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitUser)
	ctx.Conditions.SetUnknown("PermissionsReady", "Unknown")

	// Connect to the RabbitMQ server.
	rmqc, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	// Get the core data for the user from the object/context.
	username := obj.Spec.Username
	if username == "" { // TODO Switch this to a defaulting webhook.
		username = obj.Name
	}

	// Get all Permissions for a vhost, user. Add all mentioned in spec and remove unwanted.
	permissions, err := rmqc.ListPermissionsOf(username)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error listing permissions for user %s", username)
	}
	permMap := map[string]*rabbithole.PermissionInfo{}
	for _, perm := range permissions {
		permMap[perm.Vhost] = &perm
	}

	for _, perm := range obj.Spec.Permissions {
		var createPermissions, updatePermissions bool

		existingPerm, ok := permMap[perm.Vhost]
		if !ok {
			createPermissions = true
		} else {
			// Delete the entry in permMap so we can use it as a double-ended diff too.
			delete(permMap, perm.Vhost)
			updatePermissions = existingPerm.Read != perm.Read || existingPerm.Write != perm.Write || existingPerm.Configure != perm.Configure
		}

		if createPermissions || updatePermissions {
			_, err := rmqc.UpdatePermissionsIn(perm.Vhost, username, rabbithole.Permissions{
				Configure: perm.Configure,
				Read:      perm.Read,
				Write:     perm.Write,
			})
			if err != nil {
				return cu.Result{}, errors.Wrapf(err, "error updating permissions for user %s and vhost %s", username, perm.Vhost)
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
			ctx.Events.Eventf(obj, "Normal", event, "RabbitMQ permissions for user %s in vhost %s %s", username, perm.Vhost, eventMessage)
		}
	}

	//Remove any permissions that exist in RabbitMQ but not in the Spec.
	for vhost := range permMap {
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
