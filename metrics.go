package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	podresourcesapi "k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"
)

var (
	socketDir             = "/var/lib/kubelet/pod-resources"
	socketPath            = socketDir + "/kubelet.sock"
	connectionTimeout     = 10 * time.Second
	updateMetricsInterval = 2 * time.Second
)

type MetricsServer struct {
	server       *http.Server
	resourceName string
	stop         chan interface{}
	totalCount   prometheus.Gauge
	usedCount    prometheus.Gauge
}

func NewMetricsServer(port int, deviceConfig VirtualDeviceConfig) *MetricsServer {
	labels := map[string]string{
		"resourceName": deviceConfig.ResourceName,
	}
	totalCount := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "virtual_device_total_count",
		Help:        "The total count of virtual device",
		ConstLabels: labels,
	})
	totalCount.Set(float64(deviceConfig.Count))
	usedCount := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "virtual_device_used_count",
		Help:        "The used count of virtual device",
		ConstLabels: labels,
	})
	mux := http.NewServeMux()
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	mux.Handle("/metrics", promhttp.Handler())
	return &MetricsServer{
		server:       &server,
		resourceName: deviceConfig.ResourceName,
		stop:         make(chan interface{}),
		totalCount:   totalCount,
		usedCount:    usedCount,
	}
}

func (s *MetricsServer) Start() error {
	go func() {
		glog.Info("Starting metrics server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			glog.Fatalf("Failed to Listen and Server HTTP server with err: `%v`", err)
		}
	}()

	go func() {
		for {
			select {
			case <-s.stop:
				return
			default:
				err := s.updateMetrics()
				if err != nil {
					glog.Warningf("Failed to update metrics: %v", err)
				}
				time.Sleep(updateMetricsInterval)
			}
		}
	}()

	return nil
}

func (s *MetricsServer) Stop() {
	if s.server != nil {
		if err := s.server.Shutdown(context.Background()); err != nil {
			glog.Errorf("Failed to shutdown HTTP server, with err: `%v`", err)
		}
		s.server = nil
	}

	if s.stop != nil {
		close(s.stop)
		s.stop = nil
	}

	return
}

func (s *MetricsServer) updateMetrics() error {
	c, cleanup, err := connectToServer(socketPath)
	if err != nil {
		return err
	}
	defer cleanup()

	pods, err := listPods(c)
	if err != nil {
		return err
	}

	count := 0
	for _, pod := range pods.GetPodResources() {
		for _, c := range pod.Containers {
			for _, d := range c.Devices {
				if d.ResourceName == s.resourceName {
					count += len(d.DeviceIds)
				}
			}
		}
	}
	s.usedCount.Set(float64(count))

	return nil
}

func connectToServer(socket string) (*grpc.ClientConn, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, func() {}, fmt.Errorf("Failure connecting to %s: %v", socket, err)
	}

	return conn, func() { conn.Close() }, nil
}

func listPods(conn *grpc.ClientConn) (*podresourcesapi.ListPodResourcesResponse, error) {
	client := podresourcesapi.NewPodResourcesListerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("Failure getting pod resources %v", err)
	}

	return resp, nil
}
