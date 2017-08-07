/*
Copyright 2016 The Kubernetes Authors.

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
	"os"
	//"strconv"
	"testing"
	"os/exec"
	"bytes"
	//"syscall"
	"fmt"
	//"strconv"
	"strings"
	"io/ioutil"
  	"github.com/golang/glog"
)

func IsSolarFlareNICPresent() bool {

	glog.Errorf("IsSolarFlareNICPresent\n")

	var out bytes.Buffer
	var stderr bytes.Buffer

	SolarFlareNICVendorID := "1924:"

	cmdName := "lspci"
	cmd := exec.Command(cmdName, "-d", SolarFlareNICVendorID)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
	} else {
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		if (strings.Contains(out.String(), "Solarflare Communications") == true) { 
			return true
		}
	}

	return false
}

func Init() {

	glog.Errorf("Init\n");

	var out bytes.Buffer
	var stderr bytes.Buffer

	onloadver := "201606-u1.3"

	cmdName := "yum"
	cmdArgs := []string{}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
	}
	//fmt.Println("CMD--" + cmdName + ": " + out.String())

	// if yum not found, abort and return error
	if ((err != nil) && strings.Contains(stderr.String(), "not found") == false) { 
		// install onload dependencies
		cmdName = "yum"
		cmd = exec.Command(cmdName, "-y", "install", "gcc", "make", "libc", "libc-devel", "perl", "autoconf", "automake", "libtool", "kernel‐devel", "binutils", "gettext", "gawk", "gcc", "sed", "make", "bash", "glibc-common", "automake", "libtool", "libpcap", "libpcap-devel", "python-devel", "glibc‐devel.i586") 
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// unload and uninstall current onload
		cmdName = "onload_tool unload"
		cmd = exec.Command("onload_tool", "unload")
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())
		cmdName = "onload_uninstall"
		cmd = exec.Command(cmdName)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// remove current onload
		cmdName = "rm onload"
		cmd = exec.Command("/bin/sh", "-c", "rm -rf ./openonload*")
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// get open onload from a authorized source - further security todo
		cmdName = "get onload"
		cmdstring := "http://www.openonload.org/download/openonload-" + onloadver + ".tgz"
		cmd = exec.Command("wget", cmdstring) 
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// unzip onload
		cmdName = "unzip onload"
		cmdstring = "./openonload-" + onloadver + ".tgz"
		cmd = exec.Command("tar", "xvzf", cmdstring)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		//fmt.Println("CMD--" + cmdName + ": " + out.String())

		os.Chdir(os.Getenv("HOME"))
		// install current onload
		cmdName = "./openonload-" + onloadver + "/scripts/onload_install"
		cmd = exec.Command(cmdName)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
		}
		if ((err == nil) && strings.Contains(out.String(), "onload_install: Install complete")) {
			fmt.Println("CMD--" + cmdName + ": " + "Install complete")

			// reload onload
			cmdName = "onload_tool unload"
			cmd = exec.Command("onload_tool", "unload")
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
			}
			//fmt.Println("CMD--" + cmdName + ": " + out.String())
			cmdName = "onload_tool reload"
			cmd = exec.Command("onload_tool", "reload")
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
			}
			//fmt.Println("CMD--" + cmdName + ": " + out.String())

			cmdName = "onload"
			cmd = exec.Command(cmdName)
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
			}
			//fmt.Println("CMD--" + cmdName + ": " + out.String())

			if (strings.Contains(stderr.String(), "Solarflare Communications") && strings.Contains(stderr.String(), onloadver)) {
				cmdName = "/sbin/ldconfig"
				cmd = exec.Command(cmdName, "-N", "-v")
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					fmt.Println("CMD--" + cmdName + ": " + fmt.Sprint(err) + ": " + stderr.String())
				}
				//fmt.Println("CMD--" + cmdName + ": " + out.String())

				if (strings.Contains(out.String(), "libonload")) {
					if (AreAllOnloadDevicesAvailable() == true) {
						fmt.Println("Onload Install Verified\n")
					} else {
						return
					}
				} else { 
					return
				}
			}
		} else {
			return
		}

	} else {
		//Init fails with error - todo
		return
	}

	return
}

func AreAllOnloadDevicesAvailable() bool {
	glog.Errorf("AreAllOnloadDevicesAvailable\n")

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

func UnInit() {
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

func TestManagerSolarFlareNIC(t *testing.T) {

	glog.Errorf("TestManagerSolarFlareNIC\n")

	if IsSolarFlareNICPresent() == true {
		Init()
	} else {
		// clean up any exisiting device plugin software
		UnInit()
		glog.Errorf("Init aborted: no SolarFlare NICs are present\n")
	}
}
