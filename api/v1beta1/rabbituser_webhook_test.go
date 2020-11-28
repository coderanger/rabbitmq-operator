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
)

var _ = Describe("RabbitUser Webhook", func() {
	var obj *RabbitUser

	BeforeEach(func() {
		obj = &RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing", Namespace: "default"},
			Spec: RabbitUserSpec{
				Connection: RabbitConnection{
					Host:     "testhost",
					Username: "testuser",
				},
				Permissions: []RabbitPermission{
					{
						Vhost:     "/",
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				},
			},
		}
	})

	Describe("Default", func() {
		It("sets the name if unset", func() {
			obj.Default()
			Expect(obj.Spec.Username).To(Equal("testing"))
		})

		It("does not set the name if set", func() {
			obj.Spec.Username = "other"
			obj.Default()
			Expect(obj.Spec.Username).To(Equal("other"))
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

		It("rejects double permissions", func() {
			obj.Spec.Permissions = append(obj.Spec.Permissions, RabbitPermission{
				Vhost: "/",
			})
			err := obj.ValidateCreate()
			Expect(err).To(MatchError("Duplicate permissions for vhost /"))
		})

		It("rejects outputVhost with multiple permissions", func() {
			obj.Spec.OutputVhost = true
			obj.Spec.Permissions = append(obj.Spec.Permissions, RabbitPermission{
				Vhost: "other",
			})
			err := obj.ValidateCreate()
			Expect(err).To(MatchError("outputVhost can only be used with permissions for exactly one vhost"))
		})
	})
})
