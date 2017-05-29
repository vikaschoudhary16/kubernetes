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

package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	api "k8s.io/kubernetes/pkg/apis/core"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	// The reason and message set on a pod when its state cannot be confirmed as kubelet is unresponsive
	// on the node it is (was) running.
	NodeUnreachablePodReason  = "NodeLost"
	NodeUnreachablePodMessage = "Node %v which was running pod %v is unresponsive"
)

// GetHostname returns OS's hostname if 'hostnameOverride' is empty; otherwise, return 'hostnameOverride'.
func GetHostname(hostnameOverride string) string {
	hostname := hostnameOverride
	if hostname == "" {
		nodename, err := os.Hostname()
		if err != nil {
			glog.Fatalf("Couldn't determine hostname: %v", err)
		}
		hostname = nodename
	}
	return strings.ToLower(strings.TrimSpace(hostname))
}

// GetPreferredNodeAddress returns the address of the provided node, using the provided preference order.
// If none of the preferred address types are found, an error is returned.
func GetPreferredNodeAddress(node *v1.Node, preferredAddressTypes []v1.NodeAddressType) (string, error) {
	for _, addressType := range preferredAddressTypes {
		for _, address := range node.Status.Addresses {
			if address.Type == addressType {
				return address.Address, nil
			}
		}
		// If hostname was requested and no Hostname address was registered...
		if addressType == v1.NodeHostName {
			// ...fall back to the kubernetes.io/hostname label for compatibility with kubelets before 1.5
			if hostname, ok := node.Labels[kubeletapis.LabelHostname]; ok && len(hostname) > 0 {
				return hostname, nil
			}
		}
	}
	return "", fmt.Errorf("no preferred addresses found; known addresses: %v", node.Status.Addresses)
}

// GetNodeHostIP returns the provided node's IP, based on the priority:
// 1. NodeInternalIP
// 2. NodeExternalIP
func GetNodeHostIP(node *v1.Node) (net.IP, error) {
	addresses := node.Status.Addresses
	addressMap := make(map[v1.NodeAddressType][]v1.NodeAddress)
	for i := range addresses {
		addressMap[addresses[i].Type] = append(addressMap[addresses[i].Type], addresses[i])
	}
	if addresses, ok := addressMap[v1.NodeInternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	if addresses, ok := addressMap[v1.NodeExternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	return nil, fmt.Errorf("host IP unknown; known addresses: %v", addresses)
}

// InternalGetNodeHostIP returns the provided node's IP, based on the priority:
// 1. NodeInternalIP
// 2. NodeExternalIP
func InternalGetNodeHostIP(node *api.Node) (net.IP, error) {
	addresses := node.Status.Addresses
	addressMap := make(map[api.NodeAddressType][]api.NodeAddress)
	for i := range addresses {
		addressMap[addresses[i].Type] = append(addressMap[addresses[i].Type], addresses[i])
	}
	if addresses, ok := addressMap[api.NodeInternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	if addresses, ok := addressMap[api.NodeExternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	return nil, fmt.Errorf("host IP unknown; known addresses: %v", addresses)
}

// GetZoneKey is a helper function that builds a string identifier that is unique per failure-zone;
// it returns empty-string for no zone.
func GetZoneKey(node *v1.Node) string {
	labels := node.Labels
	if labels == nil {
		return ""
	}

	region, _ := labels[kubeletapis.LabelZoneRegion]
	failureDomain, _ := labels[kubeletapis.LabelZoneFailureDomain]

	if region == "" && failureDomain == "" {
		return ""
	}

	// We include the null character just in case region or failureDomain has a colon
	// (We do assume there's no null characters in a region or failureDomain)
	// As a nice side-benefit, the null character is not printed by fmt.Print or glog
	return region + ":\x00:" + failureDomain
}

// SetNodeCondition updates specific node condition with patch operation.
func SetNodeCondition(c clientset.Interface, node types.NodeName, condition v1.NodeCondition) error {
	generatePatch := func(condition v1.NodeCondition) ([]byte, error) {
		raw, err := json.Marshal(&[]v1.NodeCondition{condition})
		if err != nil {
			return nil, err
		}
		return []byte(fmt.Sprintf(`{"status":{"conditions":%s}}`, raw)), nil
	}
	condition.LastHeartbeatTime = metav1.NewTime(time.Now())
	patch, err := generatePatch(condition)
	if err != nil {
		return nil
	}
	_, err = c.CoreV1().Nodes().PatchStatus(string(node), patch)
	return err
}

// PatchNodeCIDR patches the specified node's CIDR to the given value.
func PatchNodeCIDR(c clientset.Interface, node types.NodeName, cidr string) error {
	raw, err := json.Marshal(cidr)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal CIDR: %v", err)
	}

	patchBytes := []byte(fmt.Sprintf(`{"spec":{"podCIDR":%s}}`, raw))

	if _, err := c.CoreV1().Nodes().Patch(string(node), types.StrategicMergePatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to patch node CIDR: %v", err)
	}
	return nil
}

// PatchNodeStatus patches node status.
func PatchNodeStatus(c v1core.CoreV1Interface, nodeName types.NodeName, oldNode *v1.Node, newNode *v1.Node) (*v1.Node, []byte, error) {
	patchBytes, err := preparePatchBytesforNodeStatus(nodeName, oldNode, newNode)
	if err != nil {
		return nil, nil, err
	}

	updatedNode, err := c.Nodes().Patch(string(nodeName), types.StrategicMergePatchType, patchBytes, "status")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to patch status %q for node %q: %v", patchBytes, nodeName, err)
	}
	return updatedNode, patchBytes, nil
}

func preparePatchBytesforNodeStatus(nodeName types.NodeName, oldNode *v1.Node, newNode *v1.Node) ([]byte, error) {
	oldData, err := json.Marshal(oldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal oldData for node %q: %v", nodeName, err)
	}

	// Reset spec to make sure only patch for Status or ObjectMeta is generated.
	// Note that we don't reset ObjectMeta here, because:
	// 1. This aligns with Nodes().UpdateStatus().
	// 2. Some component does use this to update node annotations.
	newNode.Spec = oldNode.Spec
	newData, err := json.Marshal(newNode)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal newData for node %q: %v", nodeName, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Node{})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateTwoWayMergePatch for node %q: %v", nodeName, err)
	}
	return patchBytes, nil
}

func ReadDeviceFiles(path string) ([]*v1.Device, error) {
	statInfo, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if statInfo.Mode().IsDir() {
		devices, err := ExtractDevicesFromDir(path)
		if err != nil {
			return devices, err
		}
		return devices, err
	} else {
		return nil, fmt.Errorf("path is not a directory")
	}
}

func ExtractDevicesFromDir(name string) ([]*v1.Device, error) {
	glog.V(1).Infof("ExtractDevicesFromDir: Entered for dir  %v", name)
	dirents, err := filepath.Glob(filepath.Join(name, "[^.]*"))
	glog.V(1).Infof("ExtractDevicesFromDir: dirents  %v", dirents)
	if err != nil {
		return nil, fmt.Errorf("glob failed: %v", err)
	}

	devices := make([]*v1.Device, 0)
	if len(dirents) == 0 {
		return devices, nil
	}

	sort.Strings(dirents)
	for _, path := range dirents {
		statInfo, err := os.Stat(path)
		if err != nil {
			glog.V(1).Infof("Can't get metadata for %q: %v", path, err)
			continue
		}

		switch {
		case statInfo.Mode().IsDir():
			glog.V(1).Infof("Not recursing into config path %q", path)
		case statInfo.Mode().IsRegular():
			device, err := ExtractDeviceFromFile(path)
			if err != nil {
				glog.V(1).Infof("Can't process config file %q: %v", path, err)
			} else {
				devices = append(devices, device)
			}
		default:
			glog.V(1).Infof("Config path %q is not a directory or file: %v", path, statInfo.Mode())
		}
	}
	return devices, nil
}

func ExtractDeviceFromFile(filename string) (device *v1.Device, err error) {
	glog.V(3).Infof("Reading config file %v", filename)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return device, err
	}

	//defaultFn := func(device *api.Device) error {
	//      return s.applyDefaults(device, filename)
	//}
	glog.V(3).Infof("Going to decode device data %v", data)

	parsed, device, deviceErr := TryDecodeSingleDevice(data)
	if parsed {
		if deviceErr != nil {
			return device, deviceErr
		}
		return device, nil
	}

	return device, fmt.Errorf("%v: read '%v', but couldn't parse as device(%v).\n",
		filename, string(data), deviceErr)
}

func TryDecodeSingleDevice(data []byte) (parsed bool, device *v1.Device, err error) {
	// JSON is valid YAML, so this should work for everything.
	glog.V(3).Infof("TryDecodeSingleDevice Entered data:%v", data)

	json, err := utilyaml.ToJSON(data)
	if err != nil {
		return false, nil, err
	}
	obj, err := runtime.Decode(api.Codecs.UniversalDecoder(), json)
	if err != nil {
		return false, device, err
	}
	glog.V(3).Infof("TryDecodeSingleDevice: getting api Device")
	// Check whether the object could be converted to single Device.
	if _, ok := obj.(*api.Device); !ok {
		err = fmt.Errorf("invalid device: %#v", obj)
		return false, device, err
	}
	newDevice := obj.(*api.Device)
	// Apply default values and validate the device.
	//if err = defaultFn(newDevice); err != nil {
	//       return true, device, err
	//}
	//if errs := validation.ValidateDevice(newDevice); len(errs) > 0 {
	//      err = fmt.Errorf("invalid device: %v", errs)
	//      return true, device, err
	//}
	glog.V(3).Infof("TryDecodeSingleDevice: getting v1 api Device")
	v1Device := &v1.Device{}
	if err := v1.Convert_api_Device_To_v1_Device(newDevice, v1Device, nil); err != nil {
		return true, nil, err
	}
	glog.V(3).Infof("TryDecodeSingleDevice: v1 api Device:%v", *v1Device)
	return true, v1Device, nil
}
