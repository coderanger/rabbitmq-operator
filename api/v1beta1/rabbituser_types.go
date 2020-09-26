/*
Copyright 2020 Noah Kantrowitz
Copyright 2019-2020 Ridecell, Inc.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RabbitmqPermission defines a single user permissions entry.
type RabbitPermission struct {
	// Vhost this applies to.
	Vhost string `json:"vhost"`
	// Configuration permissions.
	Configure string `json:"configure,omitempty"`
	// Write permissions.
	Write string `json:"write,omitempty"`
	// Read permissions.
	Read string `json:"read,omitempty"`
}

// RabbitUserSpec defines the desired state of RabbitUser
type RabbitUserSpec struct {
	Username    string             `json:"username"`
	Tags        string             `json:"tags,omitempty"`
	Permissions []RabbitPermission `json:"permissions,omitempty"`
	// TODO TopicPermissions
	Connection RabbitConnection `json:"connection,omitempty"`
}

// RabbitUserStatus defines the observed state of RabbitUser
type RabbitUserStatus struct {
}

// +kubebuilder:object:root=true

// RabbitUser is the Schema for the rabbitusers API
type RabbitUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitUserSpec   `json:"spec,omitempty"`
	Status RabbitUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitUserList contains a list of RabbitUser
type RabbitUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitUser{}, &RabbitUserList{})
}
