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

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	api "k8s.io/kubernetes/pkg/api"
)

// FakeResourceClasses implements ResourceClassInterface
type FakeResourceClasses struct {
	Fake *FakeCore
}

var resourceclassesResource = schema.GroupVersionResource{Group: "", Version: "", Resource: "resourceclasses"}

var resourceclassesKind = schema.GroupVersionKind{Group: "", Version: "", Kind: "ResourceClass"}

func (c *FakeResourceClasses) Create(resourceClass *api.ResourceClass) (result *api.ResourceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(resourceclassesResource, resourceClass), &api.ResourceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*api.ResourceClass), err
}

func (c *FakeResourceClasses) Update(resourceClass *api.ResourceClass) (result *api.ResourceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(resourceclassesResource, resourceClass), &api.ResourceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*api.ResourceClass), err
}

func (c *FakeResourceClasses) UpdateStatus(resourceClass *api.ResourceClass) (*api.ResourceClass, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(resourceclassesResource, "status", resourceClass), &api.ResourceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*api.ResourceClass), err
}

func (c *FakeResourceClasses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(resourceclassesResource, name), &api.ResourceClass{})
	return err
}

func (c *FakeResourceClasses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(resourceclassesResource, listOptions)

	_, err := c.Fake.Invokes(action, &api.ResourceClassList{})
	return err
}

func (c *FakeResourceClasses) Get(name string, options v1.GetOptions) (result *api.ResourceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(resourceclassesResource, name), &api.ResourceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*api.ResourceClass), err
}

func (c *FakeResourceClasses) List(opts v1.ListOptions) (result *api.ResourceClassList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(resourceclassesResource, resourceclassesKind, opts), &api.ResourceClassList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &api.ResourceClassList{}
	for _, item := range obj.(*api.ResourceClassList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested resourceClasses.
func (c *FakeResourceClasses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(resourceclassesResource, opts))
}

// Patch applies the patch and returns the patched resourceClass.
func (c *FakeResourceClasses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *api.ResourceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(resourceclassesResource, name, data, subresources...), &api.ResourceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*api.ResourceClass), err
}
