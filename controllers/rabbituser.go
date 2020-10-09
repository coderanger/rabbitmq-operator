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

package controllers

import (
	cu "github.com/coderanger/controller-utils"
	ctrl "sigs.k8s.io/controller-runtime"

	rabbitmqv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
	"github.com/coderanger/rabbitmq-operator/components"
	"github.com/coderanger/rabbitmq-operator/templates"
)

// +kubebuilder:rbac:groups=rabbitmq.coderanger.net,resources=rabbitusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rabbitmq.coderanger.net,resources=rabbitusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=v1,resources=secrets,verbs=get;list;watch;create;update;patch

func RabbitUser(mgr ctrl.Manager) error {
	return cu.NewReconciler(mgr).
		For(&rabbitmqv1beta1.RabbitUser{}).
		Templates(templates.Templates).
		RandomSecretComponent("RABBIT_PASSWORD").
		Component("user", components.User()).
		Component("permissions", components.Permissions()).
		TemplateComponent("user_secret.yml", "").
		ReadyStatusComponent("UserReady", "PermissionsReady").
		Webhook().
		Complete()
}
