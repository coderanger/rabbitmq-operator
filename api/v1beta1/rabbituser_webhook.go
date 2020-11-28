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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var rabbitUserLog = logf.Log.WithName("webhooks").WithName("rabbituser")

// +kubebuilder:webhook:path=/mutate-rabbitmq-coderanger-net-v1beta1-rabbituser,mutating=true,failurePolicy=fail,sideEffects=None,groups=rabbitmq.coderanger.net,resources=rabbitusers,verbs=create;update,versions=v1beta1,name=mrabbituser.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &RabbitUser{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (obj *RabbitUser) Default() {
	rabbitUserLog.Info("default", "name", obj.Name, "namespace", obj.Namespace)

	if obj.Spec.Username == "" {
		obj.Spec.Username = obj.Name
	}
}

// +kubebuilder:webhook:path=/validate-rabbitmq-coderanger-net-v1beta1-rabbituser,mutating=true,failurePolicy=fail,sideEffects=None,groups=rabbitmq.coderanger.net,resources=rabbitusers,verbs=create;update,versions=v1beta1,name=vrabbituser.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Validator = &RabbitUser{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (obj *RabbitUser) ValidateCreate() error {
	rabbitUserLog.Info("validate create", "name", obj.Name, "namespace", obj.Namespace)
	return obj.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (obj *RabbitUser) ValidateUpdate(old runtime.Object) error {
	rabbitUserLog.Info("validate update", "name", obj.Name, "namespace", obj.Namespace)
	return obj.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type. Not used, just here for interface compliance.
func (obj *RabbitUser) ValidateDelete() error {
	return nil
}

func (obj *RabbitUser) validate() error {
	// Confirm that each vhost appears only once because that's how Rabbit permissions work.
	seenVhosts := map[string]bool{}
	for _, perm := range obj.Spec.Permissions {
		_, ok := seenVhosts[perm.Vhost]
		if ok {
			return errors.Errorf("Duplicate permissions for vhost %s", perm.Vhost)
		}
		seenVhosts[perm.Vhost] = true
	}

	// Check if it's safe to use output vhost mode.
	if obj.Spec.OutputVhost && len(obj.Spec.Permissions) != 1 {
		return errors.New("outputVhost can only be used with permissions for exactly one vhost")
	}

	return nil
}
