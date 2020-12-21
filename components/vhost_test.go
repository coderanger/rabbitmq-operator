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
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("Vhost component", func() {
	var obj *rabbitv1beta1.RabbitVhost
	var rabbit *fakeRabbitClient
	var helper *cu.UnitHelper

	BeforeEach(func() {
		rabbit = newFakeRabbitClient()
		comp := Vhost()
		comp.clientFactory = rabbit.Factory
		obj = &rabbitv1beta1.RabbitVhost{
			Spec: rabbitv1beta1.RabbitVhostSpec{
				Connection: rabbitv1beta1.RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
			},
		}
		helper = suiteHelper.Setup(comp, obj)
	})

	It("creates a vhost", func() {
		helper.MustReconcile()
		Expect(rabbit.Vhosts).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("testing"),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal VhostCreated RabbitMQ vhost testing created")))
		Expect(obj).To(HaveCondition("VhostReady").WithStatus("True").WithReason("VhostExists"))
	})

	It("does not update an existing vhost with nothing to change", func() {
		rabbit.Vhosts = []*rabbithole.VhostInfo{
			{
				Name: "testing",
			},
		}
		helper.MustReconcile()
		Expect(helper.Events).ToNot(Receive())
	})
})
