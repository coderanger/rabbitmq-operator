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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("RabbitUser controller", func() {
	var helper *cu.FunctionalHelper
	// var rmqc *rabbithole.Client

	BeforeEach(func() {
		helper = suiteHelper.MustStart(RabbitUser)
		// rmqc = connect()
	})

	AfterEach(func() {
		helper.MustStop()
		helper = nil
	})

	It("runs a basic reconcile", func() {
		c := helper.TestClient

		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Tags: "management",
			},
		}
		c.Create(user)

		secret := &corev1.Secret{}
		c.EventuallyGetName("testing-password", secret)

		rmqcUser := connectUser("testing", string(secret.Data["password"]))
		Eventually(func() bool {
			_, err := rmqcUser.Whoami()
			return err == nil
		}).Should(BeTrue())
	})

	It("sets vhost permissions", func() {
		c := helper.TestClient

		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Tags: "administrator",
				Permissions: []rabbitv1beta1.RabbitPermission{
					{
						Vhost:     "/",
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				},
			},
		}
		c.Create(user)

		secret := &corev1.Secret{}
		c.EventuallyGetName("testing-password", secret)

		rmqcUser := connectUser("testing", string(secret.Data["password"]))
		Eventually(func() bool {
			_, err := rmqcUser.Whoami()
			return err == nil
		}).Should(BeTrue())

		vhosts, err := rmqcUser.ListVhosts()
		Expect(err).ToNot(HaveOccurred())
		Expect(vhosts).ToNot(BeEmpty())
	})
})
