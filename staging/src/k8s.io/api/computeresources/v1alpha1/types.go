/*
Copyright 2018 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceClass is a resource class
type ResourceClass struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines resources required
	// +optional
	Spec ResourceClassSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Most recently observed status of the resource class.
	// Populated by the system.
	// Read-only
	// +optional
	Status ResourceClassStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// Spec defines resources required
type ResourceClassSpec struct {
	// Resource Selector selects resources
	ResourceSelector []ResourcePropertySelector `json:"resourceSelector" protobuf:"bytes,1,rep,name=resourceSelector"`
	// +optional
	SubResourcesCount int32 `json:"subResourcesCount" protobuf:"varint,2,opt,name=subResourcesCount"`
}

// ResourceClassStatus  is information about the current status of a resource class
type ResourceClassStatus struct {
	Allocatable int32 `json:"allocatable" protobuf:"varint,1,name=allocatable"`
	Request     int32 `json:"request" protobuf:"varint,2,name=request"`
}

// ResourceClassList is list of rcs
type ResourceClassList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of nodes
	Items []ResourceClass `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// A resource selector operator is the set of operators that can be used in
// a resource selector requirement.
type ResourceSelectorOperator string

const (
	ResourceSelectorOpIn           ResourceSelectorOperator = "In"
	ResourceSelectorOpNotIn        ResourceSelectorOperator = "NotIn"
	ResourceSelectorOpExists       ResourceSelectorOperator = "Exists"
	ResourceSelectorOpDoesNotExist ResourceSelectorOperator = "DoesNotExist"
	ResourceSelectorOpGt           ResourceSelectorOperator = "Gt"
	ResourceSelectorOpLt           ResourceSelectorOperator = "Lt"
)

// A resource selector requirement is a selector that contains values, a key, and an operator
// that relates the key and values
type ResourceSelectorRequirement struct {
	// The label key that the selector applies to
	Key string
	// Example 0.1, intel etc
	Values []string
	// operator
	Operator ResourceSelectorOperator
}

// A null or empty selector matches no resources
type ResourcePropertySelector struct {
	// A list of resource/device selector requirements
	MatchExpressions []ResourceSelectorRequirement
}
