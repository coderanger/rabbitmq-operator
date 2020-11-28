/*
Copyright 2020 Noah Kantrowitz

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

package v1beta1

import (
	"encoding/json"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var rabbitQueueLog = logf.Log.WithName("webhooks").WithName("rabbitqueue")

// +kubebuilder:webhook:path=/mutate-rabbitmq-coderanger-net-v1beta1-rabbitqueue,mutating=true,failurePolicy=fail,sideEffects=None,groups=rabbitmq.coderanger.net,resources=rabbitqueues,verbs=create;update,versions=v1beta1,name=mrabbitqueue.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &RabbitQueue{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (obj *RabbitQueue) Default() {
	rabbitQueueLog.Info("default", "name", obj.Name, "namespace", obj.Namespace)

	if obj.Spec.QueueName == "" {
		obj.Spec.QueueName = obj.Name
	}
}

// +kubebuilder:webhook:path=/validate-rabbitmq-coderanger-net-v1beta1-rabbitqueue,mutating=true,failurePolicy=fail,sideEffects=None,groups=rabbitmq.coderanger.net,resources=rabbitqueues,verbs=create;update,versions=v1beta1,name=vrabbitqueue.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Validator = &RabbitQueue{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (obj *RabbitQueue) ValidateCreate() error {
	rabbitQueueLog.Info("validate create", "name", obj.Name, "namespace", obj.Namespace)
	return obj.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (obj *RabbitQueue) ValidateUpdate(old runtime.Object) error {
	rabbitQueueLog.Info("validate update", "name", obj.Name, "namespace", obj.Namespace)
	return obj.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type. Not used, just here for interface compliance.
func (obj *RabbitQueue) ValidateDelete() error {
	return nil
}

func (obj *RabbitQueue) validate() error {
	if obj.Spec.Arguments != nil {
		// Validate arguments.
		var args map[string]interface{}
		err := json.Unmarshal(obj.Spec.Arguments.Raw, &args)
		if err != nil {
			return errors.Wrap(err, "error parsing arguments")
		}

		// Limit values to strings, numbers, and bools.
		for key, val := range args {
			switch val.(type) {
			case string:
			case bool:
			case float64:
			default:
				return errors.Errorf("argument %s has an invalid value: %v", key, val)
			}
		}
	}
	return nil
}
