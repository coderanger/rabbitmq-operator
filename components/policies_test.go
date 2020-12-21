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
	"k8s.io/apimachinery/pkg/runtime"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("Policies component", func() {
	var obj *rabbitv1beta1.RabbitVhost
	var rabbit *fakeRabbitClient
	var helper *cu.UnitHelper

	BeforeEach(func() {
		rabbit = newFakeRabbitClient()
		comp := Policies()
		comp.clientFactory = rabbit.Factory
		obj = &rabbitv1beta1.RabbitVhost{
			Spec: rabbitv1beta1.RabbitVhostSpec{
				Connection: rabbitv1beta1.RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
				Policies: map[string]rabbitv1beta1.RabbitPolicy{},
			},
		}
		helper = suiteHelper.Setup(comp, obj)
	})

	It("creates a policy", func() {
		obj.Spec.Policies["testpol"] = rabbitv1beta1.RabbitPolicy{
			Pattern: ".*",
			Definition: runtime.RawExtension{
				Raw: []byte(`{"ha-mode": "all"}`),
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Policies).To(MatchAllKeys(Keys{
			"testing": MatchAllKeys(Keys{
				"testing-testpol": PointTo(MatchFields(IgnoreExtras, Fields{
					"Pattern": Equal(".*"),
					"Definition": MatchAllKeys(Keys{
						"ha-mode": Equal("all"),
					}),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PolicyCreated RabbitMQ policy testing-testpol for vhost testing created")))
		Expect(obj).To(HaveCondition("PoliciesReady").WithStatus("True"))
	})

	It("updates a non-matching policy", func() {
		obj.Spec.Policies["testpol"] = rabbitv1beta1.RabbitPolicy{
			Pattern: ".*",
			Definition: runtime.RawExtension{
				Raw: []byte(`{"ha-mode": "all"}`),
			},
		}
		rabbit.Policies = map[string]map[string]*rabbithole.Policy{
			"testing": {
				"testing-testpol": {
					Vhost:   "testing",
					Name:    "testing-testpol",
					Pattern: ".*",
					Definition: rabbithole.PolicyDefinition{
						"ha-mode": "exactly",
					},
				},
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Policies).To(MatchAllKeys(Keys{
			"testing": MatchAllKeys(Keys{
				"testing-testpol": PointTo(MatchFields(IgnoreExtras, Fields{
					"Pattern": Equal(".*"),
					"Definition": MatchAllKeys(Keys{
						"ha-mode": Equal("all"),
					}),
				})),
			}),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PolicyUpdated RabbitMQ policy testing-testpol for vhost testing updated")))
		Expect(obj).To(HaveCondition("PoliciesReady").WithStatus("True"))
	})

	It("deletes a policy", func() {
		rabbit.Policies = map[string]map[string]*rabbithole.Policy{
			"testing": {
				"testing-testpol": {
					Vhost:   "testing",
					Name:    "testing-testpol",
					Pattern: ".*",
					Definition: rabbithole.PolicyDefinition{
						"ha-mode": "exactly",
					},
				},
			},
		}
		helper.MustReconcile()
		Expect(rabbit.Policies).To(MatchAllKeys(Keys{
			"testing": BeEmpty(),
		}))
		Expect(helper.Events).To(Receive(Equal("Normal PolicyDeleted RabbitMQ policy testing-testpol for vhost testing deleted")))
		Expect(obj).To(HaveCondition("PoliciesReady").WithStatus("True"))
	})
})
