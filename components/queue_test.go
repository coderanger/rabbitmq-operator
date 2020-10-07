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

package components

import (
	cu "github.com/coderanger/controller-utils"
	. "github.com/coderanger/controller-utils/tests/matchers"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("Queue component", func() {
	var obj *rabbitv1beta1.RabbitQueue
	var rabbit *fakeRabbitClient
	var helper *cu.UnitHelper

	BeforeEach(func() {
		rabbit = newFakeRabbitClient()
		comp := Queue()
		comp.clientFactory = rabbit.Factory
		obj = &rabbitv1beta1.RabbitQueue{
			Spec: rabbitv1beta1.RabbitQueueSpec{
				Vhost: "/",
				Connection: rabbitv1beta1.RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
			},
		}
		helper = suiteHelper.Setup(comp, obj)
	})

	It("creates a queue", func() {
		helper.MustReconcile()
		Expect(rabbit.Queues).To(MatchAllKeys(Keys{
			"/": MatchAllKeys(Keys{
				"testing": PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("testing"),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal QueueCreated RabbitMQ queue testing on vhost / created")))
		Expect(obj).To(HaveCondition("QueueReady").WithStatus("True").WithReason("QueueExists"))
	})

	It("does not update an existing queue", func() {
		rabbit.Queues = map[string]map[string]*rabbithole.QueueInfo{
			"/": {
				"testing": {
					Name:  "testing",
					Vhost: "/",
				},
			},
		}
		helper.MustReconcile()
		Expect(helper.Events).ToNot(Receive())
	})
})
