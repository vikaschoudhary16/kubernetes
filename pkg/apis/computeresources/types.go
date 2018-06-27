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

package computeresources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceClass
type ResourceClass struct {
	metav1.TypeMeta
	// +optional
	metav1.ObjectMeta

	// Spec defines resources required
	// +optional
	Spec ResourceClassSpec
	// +optional
	Status ResourceClassStatus
}

type ResourceClassStatus struct {
	Allocatable int32
	Request     int32
}

// Spec dictates features and properties of devices targeted by Resource Class
type ResourceClassSpec struct {
	// ResourceSelector  selects resources. ORed from each selector
	ResourceSelector []ResourcePropertySelector
	// +optional
	SubResourcesCount int32
}

// RCList is a list of Rcs
type ResourceClassList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta
	Items []ResourceClass
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
