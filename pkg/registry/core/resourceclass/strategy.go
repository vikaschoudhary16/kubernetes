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

package resourceclass

import (
	"fmt"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	pkgstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/validation"
)

// resourceclassStrategy implements behavior for resourceclasses
type resourceclassStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating ResourceClass
// objects.
var Strategy = resourceclassStrategy{api.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is false for resourceclasses.
func (resourceclassStrategy) NamespaceScoped() bool {
	return false
}

// AllowCreateOnUpdate is false for resourceclasses.
func (resourceclassStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (resourceclassStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	_ = obj.(*api.ResourceClass)
	// ResourceClasss allow *all* fields, including status, to be set on create.
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (resourceclassStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	newResourceClass := obj.(*api.ResourceClass)
	oldResourceClass := old.(*api.ResourceClass)
	newResourceClass.Status = oldResourceClass.Status
}

// Validate validates a new resourceclass.
func (resourceclassStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	resourceclass := obj.(*api.ResourceClass)
	return validation.ValidateResourceClass(resourceclass)
}

// Canonicalize normalizes the object after validation.
func (resourceclassStrategy) Canonicalize(obj runtime.Object) {
}

// ValidateUpdate is the default update validation for an end user.
func (resourceclassStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	errorList := validation.ValidateResourceClass(obj.(*api.ResourceClass))
	return append(errorList, validation.ValidateResourceClassUpdate(obj.(*api.ResourceClass), old.(*api.ResourceClass))...)
}

func (ns resourceclassStrategy) Export(ctx genericapirequest.Context, obj runtime.Object, exact bool) error {
	_, ok := obj.(*api.ResourceClass)
	if !ok {
		// unexpected programmer error
		return fmt.Errorf("unexpected object: %v", obj)
	}
	ns.PrepareForCreate(ctx, obj)
	if exact {
		return nil
	}
	//ResourceClasses are the resources that allow direct status edits, therefore
	// we clear that without exact so that the resourceclass value can be reused.
	return nil
}

type resourceclassStatusStrategy struct {
	resourceclassStrategy
}

var StatusStrategy = resourceclassStatusStrategy{Strategy}

func (resourceclassStatusStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	_ = obj.(*api.ResourceClass)
	// ResourceClasss allow *all* fields, including status, to be set on create.
}

func (resourceclassStatusStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	newResourceClass := obj.(*api.ResourceClass)
	oldResourceClass := old.(*api.ResourceClass)
	newResourceClass.Spec = oldResourceClass.Spec
}

// Canonicalize normalizes the object after validation.
func (resourceclassStatusStrategy) Canonicalize(obj runtime.Object) {
}

func (resourceclassStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ResourceGetter is an interface for retrieving resources by ResourceLocation.
type ResourceGetter interface {
	Get(genericapirequest.Context, string, *metav1.GetOptions) (runtime.Object, error)
}

// ResourceClassToSelectableFields returns a field set that represents the object.
func ResourceClassToSelectableFields(resourceclass *api.ResourceClass) fields.Set {
	return generic.ObjectMetaFieldsSet(&resourceclass.ObjectMeta, false)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	resourceclassObj, ok := obj.(*api.ResourceClass)
	if !ok {
		return nil, nil, false, fmt.Errorf("not a resourceclass")
	}
	return labels.Set(resourceclassObj.ObjectMeta.Labels), ResourceClassToSelectableFields(resourceclassObj), resourceclassObj.Initializers != nil, nil
}

// MatchResourceClass returns a generic matcher for a given label and field selector.
func MatchResourceClass(label labels.Selector, field fields.Selector) pkgstorage.SelectionPredicate {
	return pkgstorage.SelectionPredicate{
		Label:       label,
		Field:       field,
		GetAttrs:    GetAttrs,
		IndexFields: []string{"metadata.name"},
	}
}
