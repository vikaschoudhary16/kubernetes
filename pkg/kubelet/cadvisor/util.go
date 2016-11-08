/*
Copyright 2015 The Kubernetes Authors.

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

package cadvisor

import (
	cadvisorApi "github.com/google/cadvisor/info/v1"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

func CapacityFromMachineInfo(info *cadvisorApi.MachineInfo) api.ResourceList {
	c := api.ResourceList{
		api.ResourceCPU: *resource.NewMilliQuantity(
			int64(info.NumCores*1000),
			resource.DecimalSI),
		api.ResourceMemory: *resource.NewQuantity(
			int64(info.MemoryCapacity),
			resource.BinarySI),
	}
	return c
}

func NumaNodeStatusListFromMachineInfo(info *cadvisorApi.MachineInfo) []api.NumaNodeStatus {
	nt := info.NumaTopology
	list_ns := make([]api.NumaNodeStatus, len(nt))
	for _, numaNode := range nt {
		capacity := api.ResourceList{
			api.ResourceCPU: *resource.NewMilliQuantity(
				// first coreof each node reserved for system
				int64(((*numaNode).NumCores-1)*1000),
				resource.DecimalSI),
			api.ResourceMemory: *resource.NewQuantity(
				int64((*numaNode).MemorySize),
				resource.BinarySI),
		}

		allocatable := api.ResourceList{
			api.ResourceCPU: *resource.NewMilliQuantity(
				// first coreof each node reserved for system
				int64(((*numaNode).NumCores-1)*1000),
				resource.DecimalSI),
			api.ResourceMemory: *resource.NewQuantity(
				int64((*numaNode).MemoryFree),
				resource.BinarySI),
		}
		ns := api.NumaNodeStatus{
			Capacity:    capacity,
			Allocatable: allocatable,
		}
		list_ns = append(list_ns, ns)
	}
	return list_ns
}
