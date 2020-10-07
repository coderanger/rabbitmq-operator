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

	cu "github.com/coderanger/controller-utils"
	"github.com/coderanger/controller-utils/randstring"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
})
