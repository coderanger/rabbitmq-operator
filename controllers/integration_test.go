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
	"github.com/coderanger/controller-utils/randstring"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("Controller integration tests", func() {
	var helper *cu.FunctionalHelper

	BeforeEach(func() {
		helper = suiteHelper.MustStart(RabbitUser, RabbitVhost, RabbitQueue)
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			helper.DebugList(&rabbitv1beta1.RabbitVhostList{})
			helper.DebugList(&rabbitv1beta1.RabbitQueueList{})
			helper.DebugList(&rabbitv1beta1.RabbitUserList{})
		}
		helper.MustStop()
		helper = nil
	})

	It("creates a vhost with a queue and user", func() {
		c := helper.TestClient

		vhost := &rabbitv1beta1.RabbitVhost{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitVhostSpec{
				VhostName: "testing-" + randstring.MustRandomString(5),
				SkipUser:  true,
			},
		}
		queue1 := &rabbitv1beta1.RabbitQueue{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitQueueSpec{
				Vhost:     vhost.Spec.VhostName,
				QueueName: "testing1-" + randstring.MustRandomString(5),
			},
		}
		queue2 := &rabbitv1beta1.RabbitQueue{
			ObjectMeta: metav1.ObjectMeta{Name: "testing2"},
			Spec: rabbitv1beta1.RabbitQueueSpec{
				Vhost:     vhost.Spec.VhostName,
				QueueName: "testing2-" + randstring.MustRandomString(5),
			},
		}
		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Username: "testing-" + randstring.MustRandomString(5),
				Tags:     "management",
				Permissions: []rabbitv1beta1.RabbitPermission{
					{
						Vhost:     vhost.Spec.VhostName,
						Read:      queue1.Spec.QueueName,
						Write:     "",
						Configure: "",
					},
				},
			},
		}
		c.Create(vhost)
		c.Create(queue1)
		c.Create(queue2)
		c.Create(user)
		c.EventuallyGetName("testing", vhost, c.EventuallyReady())
		c.EventuallyGetName("testing", queue1, c.EventuallyReady())
		c.EventuallyGetName("testing2", queue2, c.EventuallyReady())
		c.EventuallyGetName("testing", user, c.EventuallyReady())

		// Try to connect as the restricted user.
		secret := &corev1.Secret{}
		c.GetName("testing-rabbituser", secret)
		Expect(secret.Data).To(HaveKeyWithValue("RABBIT_URL_VHOST", Not(BeEmpty())))
		conn, err := amqp.Dial(string(secret.Data["RABBIT_URL_VHOST"]))
		Expect(err).ToNot(HaveOccurred())
		defer conn.Close()
		ch, err := conn.Channel()
		Expect(err).ToNot(HaveOccurred())
		defer ch.Close()

		// Try to read from testing1, which we have permissions on.
		_, err = ch.Consume(
			queue1.Spec.QueueName, // queue
			"",                    // consumer
			true,                  // auto-ack
			false,                 // exclusive
			false,                 // no-local
			false,                 // no-wait
			nil,                   // args
		)
		Expect(err).ToNot(HaveOccurred())

		// Try to read from testing2, which we do not have permissions.
		_, err = ch.Consume(
			queue2.Spec.QueueName, // queue
			"",                    // consumer
			true,                  // auto-ack
			false,                 // exclusive
			false,                 // no-local
			false,                 // no-wait
			nil,                   // args
		)
		Expect(err).To(HaveOccurred())
	})
})
