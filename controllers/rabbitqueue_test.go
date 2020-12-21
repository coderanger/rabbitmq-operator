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
	"context"
	"os"
	"strings"
	"time"

	cu "github.com/coderanger/controller-utils"
	"github.com/coderanger/controller-utils/conditions"
	"github.com/coderanger/controller-utils/randstring"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("RabbitQueue controller", func() {
	var helper *cu.FunctionalHelper
	var rmqc *rabbithole.Client

	BeforeEach(func() {
		helper = suiteHelper.MustStart(RabbitQueue)
		rmqc = connect()
	})

	AfterEach(func() {
		helper.MustStop()
		helper = nil
	})

	It("runs a basic reconcile", func() {
		c := helper.TestClient

		queue := &rabbitv1beta1.RabbitQueue{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitQueueSpec{
				Vhost:     "/",
				QueueName: "testing-" + randstring.MustRandomString(5),
			},
		}
		c.Create(queue)
		c.EventuallyGetName("testing", queue, c.EventuallyReady())
		Expect(queue.Finalizers).To(ContainElement("rabbitqueue.rabbitmq.coderanger.net/queue"))

		// Check that the queue exists
		_, err := rmqc.GetQueue("/", queue.Spec.QueueName)
		Expect(err).ToNot(HaveOccurred())

		// Delete the queue and make sure it is cleaned up.
		c.Delete(queue)
		Eventually(func() bool {
			err := helper.Client.Get(context.Background(), types.NamespacedName{Name: "testing", Namespace: helper.Namespace}, queue)
			return err != nil && kerrors.IsNotFound(err)
		}).Should(BeTrue())
		_, err = rmqc.GetQueue("/", queue.Spec.QueueName)
		Expect(err).To(HaveOccurred())
		rmqErr := err.(rabbithole.ErrorResponse)
		Expect(rmqErr.StatusCode).To(Equal(404))
	})

	It("fixes incorrect queue params", func() {
		c := helper.TestClient

		durable := true
		queue := &rabbitv1beta1.RabbitQueue{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitQueueSpec{
				Vhost:     "/",
				QueueName: "testing-" + randstring.MustRandomString(5),
				Durable:   &durable,
			},
		}

		// Create the queue without Durable.
		_, err := rmqc.DeclareQueue("/", queue.Spec.QueueName, rabbithole.QueueSettings{Durable: false})
		Expect(err).ToNot(HaveOccurred())

		c.Create(queue)
		c.EventuallyGetName("testing", queue, c.EventuallyReady())

		// Check that the queue exists
		queueInfo, err := rmqc.GetQueue("/", queue.Spec.QueueName)
		Expect(err).ToNot(HaveOccurred())
		Expect(queueInfo.Durable).To(BeTrue())

		// Put a message in the queue which should block future changes.
		amqpURI := strings.Replace(os.Getenv("TEST_RABBITMQ"), "http://", "amqp://", 1)
		amqpURI = strings.Replace(amqpURI, ":15672", ":5672", 1)
		connection, err := amqp.Dial(amqpURI)
		Expect(err).ToNot(HaveOccurred())
		channel, err := connection.Channel()
		Expect(err).ToNot(HaveOccurred())
		err = channel.Publish(
			"",
			queue.Spec.QueueName,
			true,
			false,
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte("hello world"),
				DeliveryMode:    amqp.Transient,
			},
		)
		Expect(err).ToNot(HaveOccurred())

		// Tweak the object, make sure it stays errored.
		queue.Spec.Arguments = &runtime.RawExtension{
			Raw: []byte(`{"x-max-priority": 10}`),
		}
		c.Update(queue)
		isNotReady := func() bool {
			c.Get(types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, queue)
			return conditions.IsStatusConditionPresentAndEqual(queue.Status.Conditions, "Ready", metav1.ConditionFalse)
		}
		Eventually(isNotReady).Should(BeTrue())
		Consistently(isNotReady, 2*time.Second).Should(BeTrue())
	})
})
