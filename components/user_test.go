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
	"fmt"

	cu "github.com/coderanger/controller-utils"
	. "github.com/coderanger/controller-utils/tests/matchers"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

type matchRabbitPasswordMatcher struct {
	expected string
}

func MatchRabbitPassword(password string) types.GomegaMatcher {
	return &matchRabbitPasswordMatcher{expected: password}
}

func (matcher *matchRabbitPasswordMatcher) Match(actual interface{}) (bool, error) {
	hash, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("MatchRabbitPassword matcher expects a string")
	}
	hash2, err := hashRabbitPassword(matcher.expected, rabbithole.HashingAlgorithmSHA256, hash)
	if err != nil {
		return false, err
	}
	return hash == hash2, nil
}

func (matcher *matchRabbitPasswordMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v to match the password %s", actual, matcher.expected)
}

func (matcher *matchRabbitPasswordMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v not to match the password %s", actual, matcher.expected)
}

var _ = Describe("User component", func() {
	var obj *rabbitv1beta1.RabbitUser
	var rabbit *fakeRabbitClient
	var helper *cu.UnitHelper

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
		helper.Ctx.Data["RABBIT_PASSWORD"] = "supersecret"
	})

	It("creates a user", func() {
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("testing"),
			"Tags": BeEmpty(),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal UserCreated RabbitMQ user testing created")))
		Expect(obj).To(HaveCondition("UserReady").WithStatus("False").WithReason("UserPending"))
	})

	It("creates a user only once", func() {
		helper.MustReconcile()
		Expect(helper.Events).To(Receive(Equal("Normal UserCreated RabbitMQ user testing created")))
		Expect(obj).To(HaveCondition("UserReady").WithStatus("False").WithReason("UserPending"))
		helper.MustReconcile()
		Expect(helper.Events).ToNot(Receive())
		Expect(obj).To(HaveCondition("UserReady").WithStatus("True").WithReason("UserExists"))
	})

	It("applies the Username field", func() {
		obj.Spec.Username = "other"
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("other"),
			"Tags": BeEmpty(),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal UserCreated RabbitMQ user other created")))
	})

	It("applies the Tags field", func() {
		obj.Spec.Tags = "administrator"
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("testing"),
			"Tags": ContainElements("administrator"),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal UserCreated RabbitMQ user testing created")))
	})

	It("applies the password value", func() {
		helper.Ctx.Data["RABBIT_PASSWORD"] = "extrasecret"
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name":             Equal("testing"),
			"PasswordHash":     MatchRabbitPassword("extrasecret"),
			"HashingAlgorithm": Equal(rabbithole.HashingAlgorithmSHA256),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal UserCreated RabbitMQ user testing created")))
	})

	It("does not update an existing user with nothing to change", func() {
		rabbit.Users = []*rabbithole.UserInfo{
			{
				Name:             "testing",
				PasswordHash:     "KDYrITM0cP6OZ4+ZoB0+T1SY9Ro1hbOgH4iiaPbLAAoPb0Xn", // Hash("supersecret")
				HashingAlgorithm: rabbithole.HashingAlgorithmSHA256,
			},
		}
		helper.MustReconcile()
		Expect(helper.Events).ToNot(Receive())
	})

	It("updates an existing user with the wrong password", func() {
		rabbit.Users = []*rabbithole.UserInfo{
			{
				Name:             "testing",
				PasswordHash:     "vL4eIulfhM6xfHfWRLexc8y2dmCwSuDVc2ex2FWkwmKip4kX", // Hash("other")
				HashingAlgorithm: rabbithole.HashingAlgorithmSHA256,
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name":             Equal("testing"),
			"PasswordHash":     MatchRabbitPassword("supersecret"),
			"HashingAlgorithm": Equal(rabbithole.HashingAlgorithmSHA256),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal UserUpdated RabbitMQ user testing updated")))
	})

	It("updates an existing user with the wrong tags", func() {
		obj.Spec.Tags = "monitoring"
		rabbit.Users = []*rabbithole.UserInfo{
			{
				Name:             "testing",
				Tags:             rabbithole.UserTags{"viewer"},
				PasswordHash:     "KDYrITM0cP6OZ4+ZoB0+T1SY9Ro1hbOgH4iiaPbLAAoPb0Xn", // Hash("supersecret")
				HashingAlgorithm: rabbithole.HashingAlgorithmSHA256,
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Users).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("testing"),
			"Tags": ContainElements("monitoring"),
		}))))
		Expect(helper.Events).To(Receive(Equal("Normal UserUpdated RabbitMQ user testing updated")))
	})

	It("deletes a user", func() {
		rabbit.Users = []*rabbithole.UserInfo{
			{
				Name: "testing",
			},
		}
		_, done := helper.MustFinalize()
		Expect(done).To(BeTrue())
		Expect(rabbit.Users).To(BeEmpty())
	})

	It("sets Data.vhost when permissions for only one vhost are set", func() {
		obj.Spec.Permissions = []rabbitv1beta1.RabbitPermission{
			{
				Vhost: "testing",
			},
		}
		// Twice because the first exits early after creating.
		helper.MustReconcile()
		helper.MustReconcile()
		Expect(helper.Ctx.Data).To(HaveKeyWithValue("vhost", "/testing"))
	})

	It("sets Data.vhost when permissions for only / vhost are set", func() {
		obj.Spec.Permissions = []rabbitv1beta1.RabbitPermission{
			{
				Vhost: "/",
			},
		}
		// Twice because the first exits early after creating.
		helper.MustReconcile()
		helper.MustReconcile()
		Expect(helper.Ctx.Data).To(HaveKeyWithValue("vhost", "/"))
	})

	It("does not set Data.vhost when permissions for two vhosts are set", func() {
		obj.Spec.Permissions = []rabbitv1beta1.RabbitPermission{
			{
				Vhost: "testing",
			},
			{
				Vhost: "testing2",
			},
		}
		// Twice because the first exits early after creating.
		helper.MustReconcile()
		helper.MustReconcile()
		Expect(helper.Ctx.Data).ToNot(HaveKey("vhost"))
	})

	It("does not set Data.vhost when permissions for zero vhosts are set", func() {
		obj.Spec.Permissions = []rabbitv1beta1.RabbitPermission{}
		// Twice because the first exits early after creating.
		helper.MustReconcile()
		helper.MustReconcile()
		Expect(helper.Ctx.Data).ToNot(HaveKey("vhost"))
	})

	It("does not set Data.vhost when permissions for all vhosts are set", func() {
		obj.Spec.Permissions = []rabbitv1beta1.RabbitPermission{
			{
				Vhost: "*",
			},
		}
		// Twice because the first exits early after creating.
		helper.MustReconcile()
		helper.MustReconcile()
		Expect(helper.Ctx.Data).ToNot(HaveKey("vhost"))
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
