package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/fsnotify/fsnotify"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var (
	fileDeviceDir = "/mnt/file-devices"
)

const (
	syncFileDevicePeriod = time.Minute
)

func newFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}
	return watcher, nil
}

func (m *FileDevicePlugin) syncFileDevices() error {
	fileDevices, err := listFileDevices()
	if err != nil {
		return err
	}
	m.cachedDevices = fileDevices

	go func() {
		for {
			time.Sleep(syncFileDevicePeriod)

			fileDevices, err := listFileDevices()
			if err != nil {
				log.Printf("Sync file devices failed, err: %v", err)
			}
			m.cachedDevices = fileDevices
		}
	}()
	return nil
}

func (m *FileDevicePlugin) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}
	return c, nil
}

func (m *FileDevicePlugin) deviceExists(id string) bool {
	for _, d := range m.cachedDevices {
		if d.ID == id {
			return true
		}
	}
	return false
}

func (m *FileDevicePlugin) getMounts(deviceIDs []string) (specs []*pluginapi.Mount) {
	isAllocateID := make(map[string]bool)
	for _, deviceID := range deviceIDs {
		isAllocateID[deviceID] = true
	}

	for _, d := range m.cachedDevices {
		if isAllocateID[d.ID] {
			spec := &pluginapi.Mount{
				ContainerPath: d.Path,
				HostPath:      d.Path,
			}
			specs = append(specs, spec)
		}
	}
	return specs
}

func (m *FileDevicePlugin) listDevices() (devices []*pluginapi.Device) {
	for _, fileDevice := range m.cachedDevices {
		devices = append(devices, &fileDevice.Device)
	}
	return devices
}

func listFileDevices() (fileDevices []*FileDevice, err error) {
	items, err := ioutil.ReadDir(fileDeviceDir)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if !item.IsDir() {
			fileDevice := FileDevice{}
			fileDevice.ID = item.Name()
			fileDevice.Health = pluginapi.Healthy
			fileDevice.Path = fmt.Sprintf("%v/%v", fileDeviceDir, item.Name())

			fileDevices = append(fileDevices, &fileDevice)
		}
	}
	return fileDevices, nil
}
