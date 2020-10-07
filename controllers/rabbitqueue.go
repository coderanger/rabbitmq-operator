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
)

// +kubebuilder:rbac:groups=rabbitmq.coderanger.net,resources=rabbitqueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rabbitmq.coderanger.net,resources=rabbitqueues/status,verbs=get;update;patch

func RabbitQueue(mgr ctrl.Manager) error {
	return cu.NewReconciler(mgr).
		For(&rabbitmqv1beta1.RabbitQueue{}).
		Component("queue", components.Queue()).
		ReadyStatusComponent("QueueReady").
		Webhook().
		Complete()
}
