package main

import (
	"log"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func main() {
	plugin := NewFileDevicePlugin("coldtea214/file", pluginapi.DevicePluginPath+"coldtea-file.sock")
	if err := plugin.Start(); err != nil {
		log.Fatal(err)
	}
	select {}
}
