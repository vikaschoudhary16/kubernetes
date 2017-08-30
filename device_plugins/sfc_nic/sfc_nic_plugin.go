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

package main

import (
	"os"
	//"strconv"
	"bytes"
	"os/exec"
	//"syscall"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha1"
)

const (
	onloadver    = "201606-u1.3"
	onloadsrc    = "http://www.openonload.org/download/openonload-" + onloadver + ".tgz"
	regExpSFC    = "(?m)[\r\n]+^.*SFC[6-9].*$"
	socketName   = "sfcNIC"
	resourceName = "pod.alpha.kubernetes.io/opaque-int-resource-sfcNIC"
)

// sfcNICManager manages Solarflare NIC devices
type sfcNICManager struct {
	devices     map[string]*pluginapi.Device
	deviceFiles []string
}

func NewSFCNICManager() (*sfcNICManager, error) {
	return &sfcNICManager{
		devices:     make(map[string]*pluginapi.Device),
		deviceFiles: []string{"/dev/onload", "/dev/onload_cplane", "dev/onload_epoll", "/dev/sfc_char", "/dev/sfc_affinity"},
	}, nil
}

func ExecCommand(cmdName string, arg ...string) (bytes.Buffer, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(cmdName, arg...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
	}

	return out, err
}

func (sfc *sfcNICManager) discoverSolarflareResources() bool {
	found := false
	sfc.devices = make(map[string]*pluginapi.Device)
	glog.Info("discoverSolarflareResources")
	out, err := ExecCommand("lshw", "-short", "-class", "network")
	if err != nil {
		glog.Errorf("Error while discovering: %v", err)
		return found
	}
	re := regexp.MustCompile(regExpSFC)
	sfcNICs := re.FindAllString(out.String(), -1)
	for _, nic := range sfcNICs {
		fmt.Printf("logical name: %v \n", strings.Fields(nic)[1])
		dev := pluginapi.Device{ID: strings.Fields(nic)[1], Health: pluginapi.Healthy}
		sfc.devices[strings.Fields(nic)[1]] = &dev
		found = true
	}
	fmt.Printf("Devices: %v \n", sfc.devices)

	return found
}

func (sfc *sfcNICManager) isOnloadInstallHealthy() bool {
	healthy := false
	cmdName := "onload"
	out, _ := ExecCommand(cmdName, "--version")

	if strings.Contains(out.String(), "Solarflare Communications") && strings.Contains(out.String(), onloadver) {
		cmdName = "/sbin/ldconfig"
		out, _ := ExecCommand(cmdName, "-N", "-v")

		if strings.Contains(out.String(), "libonload") {
			if AreAllOnloadDevicesAvailable() == true {
				fmt.Println("All Onload devices Verified\n")
				healthy = true
			} else {
				fmt.Errorf("Inconsistent Onload installation. All Onload devices are not available!!!")
			}
		} else {
			fmt.Errorf("Inconsistent Onload installation. libonload not detected.")
		}
	} else {
		fmt.Errorf("Inconsistent Onload installation.")
	}
	return healthy
}

func (sfc *sfcNICManager) installOnload() error {
	cmdName := "yum"
	out, err := ExecCommand(cmdName, "version")
	//fmt.Println("CMD--" + cmdName + ": " + out.String())

	// if yum not found, abort and return error
	if err == nil {
		// install onload dependencies
		cmdName = "yum"
		out, err = ExecCommand(cmdName, "-y", "install", "gcc", "make", "libc", "libc-devel", "perl", "autoconf", "automake", "libtool", "kernel‐devel", "binutils", "gettext", "gawk", "gcc", "sed", "make", "bash", "glibc-common", "automake", "libtool", "libpcap", "libpcap-devel", "python-devel", "glibc‐devel.i586", "lshw")
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// unload and uninstall current onload
		cmdName = "onload_tool unload"
		out, err = ExecCommand("onload_tool", "unload")
		//fmt.Println("CMD--" + cmdName + ": " + out.String())
		cmdName = "onload_uninstall"
		out, err = ExecCommand(cmdName)
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// remove current onload
		cmdName = "rm onload"
		out, err = ExecCommand("/bin/sh", "-c", "rm -rf ./openonload*")
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// get open onload from a authorized source - further security todo
		cmdName = "get onload"

		if strings.HasPrefix(onloadsrc, "http://") {
			out, err = ExecCommand("wget", onloadsrc)
		} else {
			out, err = ExecCommand("cp", onloadsrc, ".")
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// unzip onload
		cmdName = "unzip onload"
		cmdstring := "./openonload-" + onloadver + ".tgz"
		out, err = ExecCommand("tar", "xvzf", cmdstring)
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// install current onload
		cmdName = "./openonload-" + onloadver + "/scripts/onload_install"
		out, err = ExecCommand(cmdName)
		if err != nil {
			return fmt.Errorf("onload_install failed: %v", err)
		}
		if strings.Contains(out.String(), "onload_install: Install complete") {
			fmt.Println("CMD--" + cmdName + ": " + "Install complete")

			// reload onload
			cmdName = "onload_tool unload"
			_, err = ExecCommand("onload_tool", "unload")
			//fmt.Println("CMD--" + cmdName + ": " + out.String())
			cmdName = "onload_tool reload"
			_, err = ExecCommand("onload_tool", "reload")
			if !sfc.isOnloadInstallHealthy() {
				return fmt.Errorf("Onload Install Failed!!")
			}
		} else {
			return fmt.Errorf("onload_install could not be completed!!!!")
		}
	} else {
		return err
	}
	return nil
}

func (sfc *sfcNICManager) Init() error {
	glog.Info("Init\n")
	err := sfc.installOnload()
	return err
}

func Register(kubeletEndpoint string, pluginEndpoint, socketName string) error {
	conn, err := grpc.Dial(kubeletEndpoint, grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	defer conn.Close()
	if err != nil {
		return fmt.Errorf("device-plugin: cannot connect to kubelet service: %v", err)
	}
	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pluginEndpoint,
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return fmt.Errorf("device-plugin: cannot register to kubelet service: %v", err)
	}
	return nil
}

// Implements DevicePlugin service functions
func (sfc *sfcNICManager) ListAndWatch(emtpy *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	glog.Info("device-plugin: ListAndWatch start\n")
	for {
		sfc.discoverSolarflareResources()
		if !sfc.isOnloadInstallHealthy() {
			glog.Errorf("Error with onload installation. Marking devices unhealthy.")
			for _, device := range sfc.devices {
				device.Health = pluginapi.Unhealthy
			}
		}
		resp := new(pluginapi.ListAndWatchResponse)
		for _, dev := range sfc.devices {
			glog.Info("dev ", dev)
			resp.Devices = append(resp.Devices, dev)
		}
		glog.Info("resp.Devices ", resp.Devices)
		if err := stream.Send(resp); err != nil {
			glog.Errorf("Failed to send response to kubelet: %v\n", err)
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (sfc *sfcNICManager) Allocate(ctx context.Context, rqt *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	glog.Info("Allocate")
	resp := new(pluginapi.AllocateResponse)
	containerName := strings.Join([]string{"k8s", "POD", rqt.PodName, rqt.Namespace}, "_")
	for _, id := range rqt.DevicesIDs {
		if _, ok := sfc.devices[id]; ok {
			devRuntime := new(pluginapi.DeviceRuntimeSpec)
			for _, d := range sfc.deviceFiles {
				devRuntime.Devices = append(devRuntime.Devices, &pluginapi.DeviceSpec{
					HostPath:      d,
					ContainerPath: d,
					Permissions:   "mrw",
				})
			}
			resp.Spec = append(resp.Spec, devRuntime)
			glog.Info("Allocate interface ", id, " to ", containerName)
			go MoveInterface(containerName, id)
		}
	}
	return resp, nil
}

func MoveInterface(containerName string, interfaceName string) {
	glog.Info("Move NIC to container ", containerName, " net namespace")
	out, err := ExecCommand("/usr/bin/sfc-nic-move.sh", containerName, interfaceName)
	if err != nil {
		glog.Error(err)
	}
}

func AreAllOnloadDevicesAvailable() bool {
	glog.Info("AreAllOnloadDevicesAvailable\n")

	found := 0

	// read the whole file at once
	b, err := ioutil.ReadFile("/proc/devices")
	if err != nil {
		panic(err)
	}
	s := string(b)

	if strings.Index(s, "onload_epoll") > 0 {
		found++
	}

	if strings.Index(s, "onload_cplane") > 0 {
		found++
	}

	// '\n' is added to avoid a match with onload_cplane and onload_epoll
	if strings.Index(s, "onload\n") > 0 {
		found++
	}

	if found == 3 {
		return true
	} else {
		return false
	}
}

func (sfc *sfcNICManager) UnInit() {
	var out bytes.Buffer
	var stderr bytes.Buffer

	//fmt.Println("CMD--" + cmdName + ": " + out.String())
	cmdName := "onload_uninstall"
	cmd := exec.Command(cmdName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
	}
	//fmt.Println("CMD--" + cmdName + ": " + out.String())

	return
}

func main() {
	flag.Parse()
	fmt.Printf("Starting main \n")
	sfc, err := NewSFCNICManager()
	if err != nil {
		glog.Fatal(err)
		os.Exit(1)
	}

	found := sfc.discoverSolarflareResources()
	if !found {
		// clean up any exisiting device plugin software
		sfc.UnInit()
		glog.Errorf("Init aborted: no SolarFlare NICs are present\n")
		os.Exit(1)
	}
	if !sfc.isOnloadInstallHealthy() {
		err = sfc.Init()
		if err != nil {
			glog.Errorf("Error with onload installation: %v", err)
			for _, device := range sfc.devices {
				device.Health = pluginapi.Unhealthy
			}
		}
	}

	pluginEndpoint := fmt.Sprintf("%s-%d.sock", socketName, time.Now().Unix())
	//serverStarted := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(1)
	// Starts device plugin service.
	go func() {
		defer wg.Done()
		fmt.Printf("DveicePluginPath %s, pluginEndpoint %s\n", pluginapi.DevicePluginPath, pluginEndpoint)
		fmt.Printf("device-plugin start server at: %s\n", path.Join(pluginapi.DevicePluginPath, pluginEndpoint))
		lis, err := net.Listen("unix", path.Join(pluginapi.DevicePluginPath, pluginEndpoint))
		if err != nil {
			glog.Fatal(err)
			return
		}
		grpcServer := grpc.NewServer()
		pluginapi.RegisterDevicePluginServer(grpcServer, sfc)
		grpcServer.Serve(lis)
	}()

	// TODO: fix this
	time.Sleep(5 * time.Second)
	// Registers with Kubelet.
	err = Register(pluginapi.KubeletSocket, pluginEndpoint, resourceName)
	if err != nil {
		glog.Fatal(err)
	}
	fmt.Printf("device-plugin registered\n")
	wg.Wait()
}
