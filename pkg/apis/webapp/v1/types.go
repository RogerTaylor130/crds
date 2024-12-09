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
	Version          string `json:"VersionName"`
	Branch           string `json:"BranchName"`
	Env              string `json:"EnvName"`
	ProducerReplicas *int32 `json:"ProducerReplicas"`
	ConsumerReplicas *int32 `json:"ConsumerReplicas"`
}

// WebappStatus is the status for a Webapp resource
type WebappStatus struct {
	Consumer Consumer `json:"Consumer"`
	Producer Producer `json:"Producer"`
	Filebeat Filebeat `json:"Filebeat"`
}

type Consumer struct {
	Replicas            *int32 `json:"ConsumerReplicas"`
	AvailableReplicas   *int32 `json:"AvailableReplicas"`
	UnavailableReplicas *int32 `json:"UnavailableReplicas"`
	ObservedGeneration  *int32 `json:"ObservedGeneration"`
}

type Producer struct {
	Replicas            *int32 `json:"ConsumerReplicas"`
	AvailableReplicas   *int32 `json:"AvailableReplicas"`
	UnavailableReplicas *int32 `json:"UnavailableReplicas"`
	ObservedGeneration  *int32 `json:"ObservedGeneration"`
}

type Filebeat struct {
	Replicas            *int32 `json:"ConsumerReplicas"`
	AvailableReplicas   *int32 `json:"AvailableReplicas"`
	UnavailableReplicas *int32 `json:"UnavailableReplicas"`
	ObservedGeneration  *int32 `json:"ObservedGeneration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebappList is a Webapp of Webapp resources
type WebappList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Webapp `json:"items"`
}
