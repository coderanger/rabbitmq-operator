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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	cu "github.com/coderanger/controller-utils"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
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
	existingQueue, err := rmqc.GetQueue(vhost, queue)
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
		settings := rabbithole.QueueSettings{}
		if obj.Spec.AutoDelete != nil {
			settings.AutoDelete = *obj.Spec.AutoDelete
		}
		if obj.Spec.Durable != nil {
			settings.Durable = *obj.Spec.Durable
		}
		if obj.Spec.Arguments != nil {
			settings.Arguments = map[string]interface{}{}
			err = json.Unmarshal(obj.Spec.Arguments.Raw, &settings.Arguments)
			if err != nil {
				return cu.Result{}, errors.Wrap(err, "error parsing arguments")
			}
		}
		resp, err := rmqc.DeclareQueue(vhost, queue, settings)
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error creating queue %s on vhost %s", queue, vhost)
		}
		if resp.StatusCode != 201 {
			return cu.Result{}, errors.Errorf("unable to create queue %s on vhost %s, got response code %v", queue, vhost, resp.StatusCode)
		}
		ctx.Events.Eventf(obj, "Normal", "QueueCreated", "RabbitMQ queue %s on vhost %s created", queue, vhost)
	} else {
		// Check if the spec fields match, except we can't easily fix them if they don't since you have to drop and
		// recreate the queue and rabbithole doesn't expose the `?if-empty=true` flag on queue deletes that would make it safe(er).
		validationErrors := []string{}
		if obj.Spec.AutoDelete != nil && existingQueue.AutoDelete != *obj.Spec.AutoDelete {
			validationErrors = append(validationErrors, fmt.Sprintf("AutoDelete currently %v expecting %v", existingQueue.AutoDelete, *obj.Spec.AutoDelete))
		}
		if obj.Spec.Durable != nil && existingQueue.Durable != *obj.Spec.Durable {
			validationErrors = append(validationErrors, fmt.Sprintf("Durable currently %v expecting %v", existingQueue.Durable, *obj.Spec.Durable))
		}
		if obj.Spec.Arguments != nil {
			var args map[string]interface{}
			err = json.Unmarshal(obj.Spec.Arguments.Raw, &args)
			if err != nil {
				return cu.Result{}, errors.Wrap(err, "error parsing arguments")
			}

			for key, val := range args {
				existingVal, ok := existingQueue.Arguments[key]
				if !ok {
					validationErrors = append(validationErrors, fmt.Sprintf("Argument %s currently <not set> expecting %v", key, val))

				} else if !reflect.DeepEqual(existingVal, val) {
					validationErrors = append(validationErrors, fmt.Sprintf("Argument %s currently %v expecting %v", key, existingVal, val))
				}
			}
		}
		// Ignore any extra arguments I guess? Not sure the right behavior there.
		if len(validationErrors) != 0 {
			return cu.Result{RequeueAfter: time.Minute}, errors.Errorf("queue settings do not match: %s", strings.Join(validationErrors, ", "))
		}
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
