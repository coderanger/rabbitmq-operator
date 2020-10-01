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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/coderanger/controller-utils/conditions"
)

type RabbitPolicy struct {
	// Regular expression pattern used to match queues and exchanges,
	// , e.g. "^ha\..+"
	Pattern string `json:"pattern"`
	// What this policy applies to: "queues", "exchanges", etc.
	ApplyTo string `json:"applyTo,omitempty"`
	// Numeric priority of this policy.
	Priority int `json:"priority,omitempty"`
	// Additional arguments added to the entities (queues,
	// exchanges or both) that match a policy
	Definition map[string]runtime.RawExtension `json:"definition"`
}

// RabbitVhostSpec defines the desired state of RabbitVhost
type RabbitVhostSpec struct {
	VhostName  string                  `json:"vhostName,omitempty"`
	SkipUser   bool                    `json:"skipUser,omitempty"`
	Policies   map[string]RabbitPolicy `json:"policies,omitempty"`
	Connection RabbitConnection        `json:"connection,omitempty"`
}

// RabbitVhostStatus defines the observed state of RabbitVhost
type RabbitVhostStatus struct {
	// Represents the observations of a RabbitUsers's current state.
	// Known .status.conditions.type are: Ready, VhostReady, PoliciesReady, UserReady
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []conditions.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RabbitVhost is the Schema for the rabbitvhosts API
type RabbitVhost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitVhostSpec   `json:"spec,omitempty"`
	Status RabbitVhostStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitVhostList contains a list of RabbitVhost
type RabbitVhostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitVhost `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitVhost{}, &RabbitVhostList{})
}

// TODO code generator for this.
func (o *RabbitVhost) GetConditions() *[]conditions.Condition {
	return &o.Status.Conditions
}
