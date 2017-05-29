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

package storage

import (
	//"fmt"
	//"net/http"
	//"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/registry/cachesize"
	"k8s.io/kubernetes/pkg/registry/core/resourceclass"
)

// ResourceClassStorage includes storage for resource class and all sub resources
//type ResourceClassStorage struct {
//	Node   *REST
//	Status *StatusREST
//}

type REST struct {
	*genericregistry.Store
	//connection     client.ConnectionInfoGetter
	//proxyTransport http.RoundTripper
}

// StatusREST implements the REST endpoint for changing the status of resource class.
type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &api.ResourceClass{}
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *StatusREST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}

// NewStorage returns a ResourceClassStorage object that will work against nodes.
//func NewStorage(optsGetter generic.RESTOptionsGetter, kubeletClientConfig client.KubeletClientConfig, proxyTransport http.RoundTripper) (*NodeStorage, error) {
func NewStorage(optsGetter generic.RESTOptionsGetter) (*REST, *StatusREST) {
	store := &genericregistry.Store{
		Copier:      api.Scheme,
		NewFunc:     func() runtime.Object { return &api.ResourceClass{} },
		NewListFunc: func() runtime.Object { return &api.ResourceClassList{} },
		//PredicateFunc:     node.MatchNode,
		PredicateFunc:     resourceclass.MatchResourceClass,
		QualifiedResource: api.Resource("resourceclasses"),
		WatchCacheSize:    cachesize.GetWatchCacheSizeByResource("resourceclasses"),

		CreateStrategy: resourceclass.Strategy,
		//UpdateStrategy: resourceclass.Strategy,
		DeleteStrategy: resourceclass.Strategy,
		ExportStrategy: resourceclass.Strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: resourceclass.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err)
	}

	statusStore := *store
	//statusStore.UpdateStrategy = resourceclass.StatusStrategy

	return &REST{Store: store}, &StatusREST{store: &statusStore}
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"resourceclass"}
}
