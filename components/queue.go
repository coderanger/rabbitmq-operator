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

type queueComponent struct {
	clientFactory rabbitClientFactory
}

func Queue() *queueComponent {
	return &queueComponent{clientFactory: rabbitholeClientFactory}
}

func (_ *queueComponent) GetReadyCondition() string {
	return "QueueReady"
}

func (comp *queueComponent) Reconcile(ctx *cu.Context) (cu.Result, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitQueue)
	ctx.Conditions.SetUnknown("QueueReady", "Unknown")

	// Connect to the RabbitMQ server.
	rmqc, _, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	// Get the core data for the queue from the object/context.
	queue := obj.Spec.QueueName
	vhost := obj.Spec.Vhost

	// Check if the queue already exists. There is nothing to update since there's no secondary values (for now, maybe tracing later).
	var createQueue bool
	_, err = rmqc.GetQueue(vhost, queue)
	if err != nil {
		rabbitErr, ok := err.(rabbithole.ErrorResponse)
		if ok && rabbitErr.StatusCode == 404 {
			createQueue = true
		} else {
			return cu.Result{}, errors.Wrapf(err, "error getting queue %s on vhost %s", queue, vhost)
		}
	}

	// Create the queue if needed.
	if createQueue {
		resp, err := rmqc.DeclareQueue(vhost, queue, rabbithole.QueueSettings{})
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error creating queue %s on vhost %s", queue, vhost)
		}
		if resp.StatusCode != 201 {
			return cu.Result{}, errors.Errorf("unable to create queue %s on vhost %s, got response code %v", queue, vhost, resp.StatusCode)
		}
		ctx.Events.Eventf(obj, "Normal", "QueueCreated", "RabbitMQ queue %s on vhost %s created", queue, vhost)
	}

	ctx.Conditions.SetfTrue("QueueReady", "QueueExists", "RabbitMQ queue %s on vhost %s exists", queue, vhost)
	return cu.Result{}, nil
}

func (comp *queueComponent) Finalize(ctx *cu.Context) (cu.Result, bool, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitQueue)

	// Connect to the RabbitMQ server.
	rmqc, _, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, false, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	_, err = rmqc.DeleteQueue(obj.Spec.Vhost, obj.Spec.QueueName)
	if err != nil {
		return cu.Result{}, false, errors.Wrapf(err, "error deleting rabbitmq queue %s on vhost %s", obj.Spec.QueueName, obj.Spec.Vhost)
	}
	return cu.Result{}, true, nil
}
