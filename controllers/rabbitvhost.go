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

// +kubebuilder:rbac:groups=rabbitmq.coderanger.net,resources=rabbitvhosts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rabbitmq.coderanger.net,resources=rabbitvhosts/status,verbs=get;update;patch

func RabbitVhost(mgr ctrl.Manager) error {
	return cu.NewReconciler(mgr).
		For(&rabbitmqv1beta1.RabbitVhost{}).
		Templates(templates.Templates).
		Component("vhost", components.Vhost()).
		Component("policies", components.Policies()).
		TemplateComponent("vhost_user.yml", "UserReady").
		ReadyStatusComponent("VhostReady", "PoliciesReady", "UserReady").
		Webhook().
		Complete()
}
