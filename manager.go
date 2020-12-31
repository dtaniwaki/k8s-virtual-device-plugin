package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var (
	devicePluginPath = os.Getenv("DEVICE_PLUGIN_PATH")
)

func init() {
	if devicePluginPath == "" {
		devicePluginPath = pluginapi.DevicePluginPath
	}
}

// VirtualDeviceManager manages our virtual devices
type VirtualDeviceManager struct {
	resourceName string
	socketName   string
	devices      map[string]*pluginapi.Device
	server       *grpc.Server
	health       chan *pluginapi.Device
}

var _ pluginapi.DevicePluginServer = &VirtualDeviceManager{}

func NewVirtualDeviceManager(devConfig VirtualDeviceConfig) (*VirtualDeviceManager, error) {
	vdm := &VirtualDeviceManager{
		devices:      map[string]*pluginapi.Device{},
		resourceName: devConfig.ResourceName,
		socketName:   devConfig.SocketName,
		health:       make(chan *pluginapi.Device),
	}

	for i := 1; i <= devConfig.Count; i++ {
		resourceName := fmt.Sprintf("%s-%d", devConfig.ResourceName, i)
		newDev := pluginapi.Device{ID: resourceName, Health: pluginapi.Healthy}
		vdm.devices[resourceName] = &newDev
	}

	return vdm, nil
}

// Start starts the gRPC server of the device plugin
func (vdm *VirtualDeviceManager) Start() error {
	err := vdm.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", vdm.socketPath())
	if err != nil {
		return err
	}

	vdm.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(vdm.server, vdm)

	go vdm.server.Serve(sock)

	// Wait for server to start by launching a blocking connection
	conn, err := grpc.Dial(vdm.socketPath(), grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return err
	}

	conn.Close()

	go vdm.healthcheck()

	return nil
}

// Stop stops the gRPC server
func (vdm *VirtualDeviceManager) Stop() {
	if vdm.server != nil {
		vdm.server.Stop()
		vdm.server = nil
	}

	err := vdm.cleanup()
	if err != nil {
		glog.Errorf("Failed to clean up: `%v`", err)
	}
}

// healthcheck monitors and updates device status
// TODO: Currently does nothing, need to implement
func (vdm *VirtualDeviceManager) healthcheck() error {
	for {
		glog.Info(vdm.devices)
		time.Sleep(60 * time.Second)
	}
}

func (vdm *VirtualDeviceManager) cleanup() error {
	if err := os.Remove(vdm.socketPath()); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (vdm *VirtualDeviceManager) socketPath() string {
	return path.Join(devicePluginPath, vdm.socketName)
}

// Register with kubelet
func (vdm *VirtualDeviceManager) Register() error {
	conn, err := grpc.Dial(pluginapi.KubeletSocket, grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	defer conn.Close()
	if err != nil {
		return fmt.Errorf("device-plugin: cannot connect to kubelet service: %v", err)
	}
	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version: pluginapi.Version,
		// Name of the unix socket the device plugin is listening on
		// PATH = path.Join(DevicePluginPath, endpoint)
		Endpoint: vdm.socketName,
		// Schedulable resource name.
		ResourceName: vdm.resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return fmt.Errorf("device-plugin: cannot register to kubelet service: %v", err)
	}
	return nil
}

func (vdm *VirtualDeviceManager) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

// ListAndWatch lists devices and update that list according to the health status
func (vdm *VirtualDeviceManager) ListAndWatch(e *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	glog.Info("device-plugin: ListAndWatch start")
	resp := new(pluginapi.ListAndWatchResponse)
	for _, dev := range vdm.devices {
		glog.Info("dev ", dev)
		resp.Devices = append(resp.Devices, dev)
	}
	glog.Info("resp.Devices ", resp.Devices)
	if err := stream.Send(resp); err != nil {
		glog.Errorf("Failed to send response to kubelet: %v", err)
	}

	for {
		select {
		case d := <-vdm.health:
			d.Health = pluginapi.Unhealthy
			resp := new(pluginapi.ListAndWatchResponse)
			for _, dev := range vdm.devices {
				glog.Info("dev ", dev)
				resp.Devices = append(resp.Devices, dev)
			}
			glog.Info("resp.Devices ", resp.Devices)
			if err := stream.Send(resp); err != nil {
				glog.Errorf("Failed to send response to kubelet: %v", err)
			}
		}
	}
}

// Allocate devices
func (vdm *VirtualDeviceManager) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	glog.Info("Allocate")
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		for _, id := range req.DevicesIDs {
			if _, ok := vdm.devices[id]; !ok {
				glog.Errorf("Can't allocate interface %s", id)
				return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
			}
		}
		glog.Info("Allocated interfaces ", req.DevicesIDs)
		envresourceName := fmt.Sprintf("VIRTUAL_DEVICE_%s", strings.ReplaceAll(strings.ToUpper(vdm.resourceName), "-", "_"))
		annotationresourceName := fmt.Sprintf("virtual-device/%s", vdm.resourceName)
		deviceIDsStr := strings.Join(req.DevicesIDs, ",")
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				envresourceName: deviceIDsStr,
			},
			Annotations: map[string]string{
				annotationresourceName: deviceIDsStr,
			},
		}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}
	return &responses, nil
}

func (m *VirtualDeviceManager) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}
