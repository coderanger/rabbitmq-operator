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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/coderanger/controller-utils/conditions"
)

// RabbitUserSpec defines the desired state of RabbitUser
type RabbitQueueSpec struct {
	QueueName  string `json:"queueName,omitempty"`
	Vhost      string `json:"vhost"`
	AutoDelete *bool  `json:"autoDelete,omitempty"`
	Durable    *bool  `json:"durable,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Arguments  *runtime.RawExtension `json:"arguments,omitempty"`
	Connection RabbitConnection      `json:"connection,omitempty"`
}

// RabbitQueueStatus defines the observed state of RabbitQueue
type RabbitQueueStatus struct {
	// Represents the observations of a RabbitQueues's current state.
	// Known .status.conditions.type are: Ready, QueueReady
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []conditions.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RabbitQueue is the Schema for the rabbitQueues API
type RabbitQueue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitQueueSpec   `json:"spec,omitempty"`
	Status RabbitQueueStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitQueueList contains a list of RabbitQueue
type RabbitQueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitQueue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitQueue{}, &RabbitQueueList{})
}

// TODO code generator for this.
func (o *RabbitQueue) GetConditions() *[]conditions.Condition {
	return &o.Status.Conditions
}
