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

package deviceplugin

import (
	"sync"

	"google.golang.org/grpc"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha1"
)

// MonitorCallback is the function called when a device's health state changes,
// or new devices are reported, or old devices are deleted.
// Updated contains the most recent state of the Device.
type MonitorCallback func(resourceName string, added, updated, deleted []*pluginapi.Device)

// Manager manages all the Device Plugins running on a node.
type Manager interface {
	// Start starts the gRPC Registration service.
	Start() error
	// Devices is the map of devices that have registered themselves
	// against the manager.
	// The map key is the ResourceName of the device plugins.
	Devices() map[string][]*pluginapi.Device

	// Allocate calls the gRPC Allocate on the device plugin.
	Allocate(string, string, string, []string) (*pluginapi.AllocateResponse, error)

	// Stop stops the manager.
	Stop() error
}

// ManagerImpl is the structure in charge of managing Device Plugins.
type ManagerImpl struct {
	socketname string
	socketdir  string

	Endpoints map[string]*endpoint // Key is ResourceName
	mutex     sync.Mutex

	callback MonitorCallback

	server *grpc.Server
}

const (
	// errDevicePluginUnknown is the error raised when the device Plugin returned by Monitor is not know by the Device Plugin manager
	errDevicePluginUnknown = "Manager does not have device plugin for device:"
	// errDeviceUnknown is the error raised when the device returned by Monitor is not know by the Device Plugin manager
	errDeviceUnknown = "Could not find device in it's Device Plugin's Device List:"
	// errBadSocket is the error raised when the registry socket path is not absolute
	errBadSocket = "Bad socketPath, must be an absolute path:"
	// errRemoveSocket is the error raised when the registry could not remove the existing socket
	errRemoveSocket = "Failed to remove socket while starting device plugin registry, with error"
	// errListenSocket is the error raised when the registry could not listen on the socket
	errListenSocket = "Failed to listen to socket while starting device plugin registry, with error"
	// errListAndWatch is the error raised when ListAndWatch ended unsuccessfully
	errListAndWatch = "ListAndWatch ended unexpectedly for device plugin %s with error %v"
)
