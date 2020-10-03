/*
Copyright 2020 Noah Kantrowitz
Copyright 2018 Ridecell, Inc.

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

type vhostComponent struct {
	clientFactory rabbitClientFactory
}

func Vhost() *vhostComponent {
	return &vhostComponent{clientFactory: rabbitholeClientFactory}
}

func (comp *vhostComponent) Reconcile(ctx *cu.Context) (cu.Result, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitVhost)
	ctx.Conditions.SetUnknown("VhostReady", "Unknown")

	// Connect to the RabbitMQ server.
	rmqc, _, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	// Get the core data for the vhost from the object/context.
	vhost := obj.Spec.VhostName

	// Check if the vhost already exists. There is nothing to update since there's no secondary values (for now, maybe tracing later).
	var createVhost bool
	_, err = rmqc.GetVhost(vhost)
	if err != nil {
		rabbitErr, ok := err.(rabbithole.ErrorResponse)
		if ok && rabbitErr.StatusCode == 404 {
			createVhost = true
		} else {
			return cu.Result{}, errors.Wrapf(err, "error getting vhost %s", vhost)
		}
	}

	// Create the vhost if needed.
	if createVhost {
		resp, err := rmqc.PutVhost(vhost, rabbithole.VhostSettings{})
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error creating vhost %s", vhost)
		}
		if resp.StatusCode != 201 {
			return cu.Result{}, errors.Errorf("unable to create vhost %s, got response code %v", vhost, resp.StatusCode)
		}
		ctx.Events.Eventf(obj, "Normal", "VhostCreated", "RabbitMQ vhost %s created", vhost)
	}

	ctx.Conditions.SetfTrue("VhostReady", "VhostExists", "RabbitMQ vhost %s exists", vhost)
	return cu.Result{}, nil
}
