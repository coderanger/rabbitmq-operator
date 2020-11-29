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
var rabbitVhostLog = logf.Log.WithName("webhooks").WithName("rabbitvhost")

// +kubebuilder:webhook:path=/mutate-rabbitmq-coderanger-net-v1beta1-rabbitvhost,mutating=true,failurePolicy=fail,sideEffects=None,groups=rabbitmq.coderanger.net,resources=rabbitvhosts,verbs=create;update,versions=v1beta1,name=mrabbitvhost.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &RabbitVhost{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (obj *RabbitVhost) Default() {
	rabbitVhostLog.Info("default", "name", obj.Name, "namespace", obj.Namespace)

	if obj.Spec.VhostName == "" {
		obj.Spec.VhostName = obj.Name
	}
}

// +kubebuilder:webhook:path=/validate-rabbitmq-coderanger-net-v1beta1-rabbitvhost,mutating=true,failurePolicy=fail,sideEffects=None,groups=rabbitmq.coderanger.net,resources=rabbitvhosts,verbs=create;update,versions=v1beta1,name=vrabbitvhost.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Validator = &RabbitVhost{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (obj *RabbitVhost) ValidateCreate() error {
	rabbitVhostLog.Info("validate create", "name", obj.Name, "namespace", obj.Namespace)
	return obj.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (obj *RabbitVhost) ValidateUpdate(old runtime.Object) error {
	rabbitVhostLog.Info("validate update", "name", obj.Name, "namespace", obj.Namespace)
	return obj.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type. Not used, just here for interface compliance.
func (obj *RabbitVhost) ValidateDelete() error {
	return nil
}

func (obj *RabbitVhost) validate() error {
	// Validate policies.
	for name, specPolicy := range obj.Spec.Policies {
		var definition map[string]interface{}
		err := json.Unmarshal(specPolicy.Definition.Raw, &definition)
		if err != nil {
			return errors.Wrapf(err, "error parsing definition %s", name)
		}
		for key, val := range definition {
			switch key {
			case "ha-mode":
				strVal, ok := val.(string)
				if !ok {
					return errors.Errorf("policy %s ha-mode value is not a string: %#v", name, val)
				}
				if strVal != "exactly" && strVal != "all" && strVal != "nodes" {
					return errors.Errorf("policy %s ha-mode value is not a known HA mode: %s", name, strVal)
				}
			// TODO More validation for common keys.
			default:
				_, isStr := val.(string)
				_, isBool := val.(bool)
				_, isNum := val.(float64)
				if !isStr && !isBool && !isNum {
					return errors.Errorf("policy %s %s value is not a string, boolean, or number: %#v", name, key, val)
				}
			}
		}
	}
	return nil
}
