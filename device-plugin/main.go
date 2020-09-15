package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func main() {
	plugin := NewFileDevicePlugin("coldtea214/file", pluginapi.DevicePluginPath+"coldtea-file.sock")
	if err := plugin.Start(); err != nil {
		log.Fatal(err)
	}

	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Fatal("Failed to created FS watcher.")
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				plugin.Stop()
				if err := plugin.Start(); err != nil {
					log.Fatal(err)
				}
			}
		case err := <-watcher.Errors:
			log.Printf("inotify: %s", err)
		}
	}
}
