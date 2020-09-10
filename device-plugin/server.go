package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type FileDevice struct {
	pluginapi.Device
	Path string
}

type FileDevicePlugin struct {
	resourceName  string
	socket        string
	server        *grpc.Server
	cachedDevices []*FileDevice
}

func NewFileDevicePlugin(resourceName string, socket string) *FileDevicePlugin {
	return &FileDevicePlugin{
		resourceName: resourceName,
		socket:       socket,
		server:       grpc.NewServer([]grpc.ServerOption{}...),
	}
}

func (m *FileDevicePlugin) Start() error {
	if err := m.syncFileDevices(); err != nil {
		log.Printf("Could not init device plugin for '%s': %s", m.resourceName, err)
	}

	if err := m.Serve(); err != nil {
		log.Printf("Could not start device plugin for '%s': %s", m.resourceName, err)
		return err
	}
	log.Printf("Starting to serve '%s' on %s", m.resourceName, m.socket)

	if err := m.Register(); err != nil {
		log.Printf("Could not register device plugin: %s", err)
		return err
	}
	log.Printf("Registered device plugin for '%s' with Kubelet", m.resourceName)

	return nil
}

func (m *FileDevicePlugin) Serve() error {
	os.Remove(m.socket)
	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	pluginapi.RegisterDevicePluginServer(m.server, m)

	go func() {
		log.Printf("Starting GRPC server for '%s'", m.resourceName)
		if err := m.server.Serve(sock); err != nil {
			log.Fatalf("GRPC server for '%s' crashed with error: %v", m.resourceName, err)
		}
	}()

	// 确认 grpc server 是否启动成功
	conn, err := m.dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

func (m *FileDevicePlugin) Register() error {
	conn, err := m.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: m.resourceName,
	}

	_, err = client.Register(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}

func (m *FileDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return nil, nil
}

func (m *FileDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	log.Printf("List response: %v", m.listDevices())
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.listDevices()})

	for {
		time.Sleep(syncFileDevicePeriod)
		log.Printf("Watch response: %v", m.listDevices())
		s.Send(&pluginapi.ListAndWatchResponse{Devices: m.listDevices()})
	}
}

func (m *FileDevicePlugin) GetPreferredAllocation(ctx context.Context, r *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}

func (m *FileDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		for _, id := range req.DevicesIDs {
			if !m.deviceExists(id) {
				log.Printf("Device %v is invalid", id)
				return nil, fmt.Errorf("invalid allocation request for '%s': unknown device: %s", m.resourceName, id)
			}
		}

		response := pluginapi.ContainerAllocateResponse{}
		response.Mounts = m.getMounts(req.DevicesIDs)
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}
	log.Printf("Allocate response: %+v", responses.ContainerResponses)
	return &responses, nil
}

func (m *FileDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return nil, nil
}
