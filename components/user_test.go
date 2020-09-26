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
	rabbithole "github.com/michaelklishin/rabbit-hole"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/coderanger/controller-utils/tests"
	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("User component", func() {
	var obj *rabbitv1beta1.RabbitUser
	var rabbit *fakeRabbitClient
	var helper *tests.UnitHelper

	BeforeEach(func() {
		rabbit = newFakeRabbitClient()
		comp := User()
		comp.clientFactory = rabbit.Factory
		obj = &rabbitv1beta1.RabbitUser{
			Spec: rabbitv1beta1.RabbitUserSpec{
				Connection: rabbitv1beta1.RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
			},
		}
		helper = suiteHelper.Setup(comp, obj)
		helper.Ctx.Data["password"] = "supersecret"
	})

	It("creates a user", func() {
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("testing"),
			"Tags": Equal(""),
		})))
		Expect(helper.Events).To(Receive(Equal("Normal Created RabbitMQ user testing created")))
	})

	It("applies the Username field", func() {
		obj.Spec.Username = "other"
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("other"),
			"Tags": Equal(""),
		})))
		Expect(helper.Events).To(Receive(Equal("Normal Created RabbitMQ user other created")))
	})

	It("applies the Tags field", func() {
		obj.Spec.Tags = "administrator"
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("testing"),
			"Tags": Equal("administrator"),
		})))
		Expect(helper.Events).To(Receive(Equal("Normal Created RabbitMQ user testing created")))
	})
})

var _ = Describe("hashRabbitPassword", func() {
	It("can verify a hash produced by RabbitMQ itself", func() {
		// Test hash from the default rabbitmq docker image.
		existingHash := "xCVmNzPVKZ+UFbHtacLM6/3f/atNng4nD8L2koLijbqJqoyF"
		hash, err := hashRabbitPassword("guest", rabbithole.HashingAlgorithmSHA256, existingHash)
		Expect(err).ToNot(HaveOccurred())
		Expect(hash).To(Equal(existingHash))
	})
})
