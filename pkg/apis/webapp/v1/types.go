/*
Copyright 2017 The Kubernetes Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Webapp is a specification for a Webapp resource
type Webapp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebappSpec   `json:"spec"`
	Status WebappStatus `json:"status"`
}

// WebappSpec is the spec for a Webapp resource
type WebappSpec struct {
	Version          string `json:"version"`
	Branch           string `json:"branch"`
	Env              string `json:"env"`
	ProducerReplicas *int32 `json:"producerReplicas"`
	ConsumerReplicas *int32 `json:"consumerReplicas"`
}

// WebappStatus is the status for a Webapp resource
type WebappStatus struct {
	Consumer Consumer `json:"consumer"`
	Producer Producer `json:"producer"`
	Filebeat Filebeat `json:"filebeat"`
}

type Consumer struct {
	Replicas            *int32 `json:"consumerReplicas"`
	AvailableReplicas   *int32 `json:"consumerAvailableReplicas"`
	UnavailableReplicas *int32 `json:"consumerUnavailableReplicas"`
	ObservedGeneration  *int32 `json:"consumerObservedGeneration"`
}

type Producer struct {
	Replicas            *int32 `json:"producerReplicas"`
	AvailableReplicas   *int32 `json:"producerAvailableReplicas"`
	UnavailableReplicas *int32 `json:"producerUnavailableReplicas"`
	ObservedGeneration  *int32 `json:"producerObservedGeneration"`
}

type Filebeat struct {
	Replicas            *int32 `json:"filebeatReplicas"`
	AvailableReplicas   *int32 `json:"filebeatAvailableReplicas"`
	UnavailableReplicas *int32 `json:"filebeatUnavailableReplicas"`
	ObservedGeneration  *int32 `json:"filebeatObservedGeneration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebappList is a Webapp of Webapp resources
type WebappList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Webapp `json:"items"`
}
