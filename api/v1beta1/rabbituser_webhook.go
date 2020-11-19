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
