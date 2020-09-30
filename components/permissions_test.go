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

var _ = Describe("Permissions component", func() {
	var obj *rabbitv1beta1.RabbitUser
	var rabbit *fakeRabbitClient
	var helper *cu.UnitHelper

	BeforeEach(func() {
		rabbit = newFakeRabbitClient()
		comp := Permissions()
		comp.clientFactory = rabbit.Factory
		obj = &rabbitv1beta1.RabbitUser{
			Spec: rabbitv1beta1.RabbitUserSpec{
				Connection: rabbitv1beta1.RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
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
		helper = suiteHelper.Setup(comp, obj)
	})

	It("creates a permission", func() {
		helper.MustReconcile()
		Expect(rabbit.Permissions).To(MatchAllKeys(Keys{
			"testing": MatchAllKeys(Keys{
				"/": PointTo(MatchFields(IgnoreExtras, Fields{
					"Read":      Equal(".*"),
					"Write":     Equal(".*"),
					"Configure": Equal(".*"),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PermissionsCreated RabbitMQ permissions for user testing in vhost / created")))
		Expect(obj).To(HaveCondition("PermissionsReady").WithStatus("True"))
	})

	It("updates a permission when Read does not match", func() {
		rabbit.Permissions = map[string]map[string]*rabbithole.PermissionInfo{
			"testing": {
				"/": {
					User:      "testing",
					Vhost:     "/",
					Read:      "other",
					Write:     ".*",
					Configure: ".*",
				},
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Permissions).To(MatchAllKeys(Keys{
			"testing": MatchAllKeys(Keys{
				"/": PointTo(MatchFields(IgnoreExtras, Fields{
					"Read":      Equal(".*"),
					"Write":     Equal(".*"),
					"Configure": Equal(".*"),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PermissionsUpdated RabbitMQ permissions for user testing in vhost / updated")))
	})

	It("updates a permission when Write does not match", func() {
		rabbit.Permissions = map[string]map[string]*rabbithole.PermissionInfo{
			"testing": {
				"/": {
					User:      "testing",
					Vhost:     "/",
					Read:      ".*",
					Write:     "other",
					Configure: ".*",
				},
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Permissions).To(MatchAllKeys(Keys{
			"testing": MatchAllKeys(Keys{
				"/": PointTo(MatchFields(IgnoreExtras, Fields{
					"Read":      Equal(".*"),
					"Write":     Equal(".*"),
					"Configure": Equal(".*"),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PermissionsUpdated RabbitMQ permissions for user testing in vhost / updated")))
	})

	It("updates a permission when Configure does not match", func() {
		rabbit.Permissions = map[string]map[string]*rabbithole.PermissionInfo{
			"testing": {
				"/": {
					User:      "testing",
					Vhost:     "/",
					Read:      ".*",
					Write:     ".*",
					Configure: "other",
				},
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Permissions).To(MatchAllKeys(Keys{
			"testing": MatchAllKeys(Keys{
				"/": PointTo(MatchFields(IgnoreExtras, Fields{
					"Read":      Equal(".*"),
					"Write":     Equal(".*"),
					"Configure": Equal(".*"),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PermissionsUpdated RabbitMQ permissions for user testing in vhost / updated")))
	})

	It("does not update any permissions when all values match", func() {
		rabbit.Permissions = map[string]map[string]*rabbithole.PermissionInfo{
			"testing": {
				"/": {
					User:      "testing",
					Vhost:     "/",
					Read:      ".*",
					Write:     ".*",
					Configure: ".*",
				},
			},
		}
		helper.MustReconcile()
		Expect(helper.Events).ToNot(Receive())
	})

	It("deletes a permission not in the spec", func() {
		rabbit.Permissions = map[string]map[string]*rabbithole.PermissionInfo{
			"testing": {
				"/": {
					User:      "testing",
					Vhost:     "/",
					Read:      ".*",
					Write:     ".*",
					Configure: ".*",
				},
			},
		}
		obj.Spec.Permissions = []rabbitv1beta1.RabbitPermission{}
		helper.MustReconcile()
		Expect(rabbit.Permissions).To(BeEmpty())
		Expect(helper.Events).To(Receive(Equal("Normal PermissionsDeleted RabbitMQ permissions for user testing in vhost / deleted")))
	})
})
