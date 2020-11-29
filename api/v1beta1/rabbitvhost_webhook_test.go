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

package v1beta1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("RabbitVhost Webhook", func() {
	var obj *RabbitVhost

	BeforeEach(func() {
		obj = &RabbitVhost{
			ObjectMeta: metav1.ObjectMeta{Name: "testing", Namespace: "default"},
			Spec: RabbitVhostSpec{
				Connection: RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
				Policies: map[string]RabbitPolicy{},
			},
		}
	})

	Describe("Default", func() {
		It("sets the name if unset", func() {
			obj.Default()
			Expect(obj.Spec.VhostName).To(Equal("testing"))
		})

		It("does not set the name if set", func() {
			obj.Spec.VhostName = "other"
			obj.Default()
			Expect(obj.Spec.VhostName).To(Equal("other"))
		})
	})

	Describe("Validate", func() {
		It("accepts a simple object", func() {
			err := obj.ValidateCreate()
			Expect(err).ToNot(HaveOccurred())
			err = obj.ValidateUpdate(obj)
			Expect(err).ToNot(HaveOccurred())
			err = obj.ValidateDelete()
			Expect(err).ToNot(HaveOccurred())
		})

		It("rejects a malformed policy", func() {
			obj.Spec.Policies["bad"] = RabbitPolicy{
				Pattern: ".*",
				Definition: runtime.RawExtension{
					Raw: []byte(`{`),
				},
			}
			err := obj.ValidateCreate()
			Expect(err).To(MatchError("error parsing definition bad: unexpected end of JSON input"))
		})

		It("rejects a malformed ha-mode", func() {
			obj.Spec.Policies["ha"] = RabbitPolicy{
				Pattern: ".*",
				Definition: runtime.RawExtension{
					Raw: []byte(`{"ha-mode": "qwer"}`),
				},
			}
			err := obj.ValidateCreate()
			Expect(err).To(MatchError("policy ha ha-mode value is not a known HA mode: qwer"))
		})

		It("accepts an all ha-mode", func() {
			obj.Spec.Policies["ha"] = RabbitPolicy{
				Pattern: ".*",
				Definition: runtime.RawExtension{
					Raw: []byte(`{"ha-mode": "all"}`),
				},
			}
			err := obj.ValidateCreate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("accepts an exactly ha-mode", func() {
			obj.Spec.Policies["ha"] = RabbitPolicy{
				Pattern: ".*",
				Definition: runtime.RawExtension{
					Raw: []byte(`{"ha-mode": "exactly"}`),
				},
			}
			err := obj.ValidateCreate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("accepts an nodes ha-mode", func() {
			obj.Spec.Policies["ha"] = RabbitPolicy{
				Pattern: ".*",
				Definition: runtime.RawExtension{
					Raw: []byte(`{"ha-mode": "nodes"}`),
				},
			}
			err := obj.ValidateCreate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("rejects a malformed definition key", func() {
			obj.Spec.Policies["bad"] = RabbitPolicy{
				Pattern: ".*",
				Definition: runtime.RawExtension{
					Raw: []byte(`{"asdf": []}`),
				},
			}
			err := obj.ValidateCreate()
			Expect(err).To(MatchError("policy bad asdf value is not a string, boolean, or number: []interface {}{}"))
		})
	})
})
