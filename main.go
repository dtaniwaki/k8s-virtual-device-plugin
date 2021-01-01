package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtaniwaki/k8s-virtual-device-plugin/pkg"

	"github.com/golang/glog"
)

var (
	Version      string = "Unknown"
	Revision     string = "Unknown"
	GitTagState  string = "Unknown"
	GitTreeState string = "Unknown"
)

func main() {
	noRegister := flag.Bool("no-register", false, "Run without registering to kubelet")
	metricsPort := flag.Int("metrics-port", 2112, "Metrics port")
	flag.Parse()
	err := flag.Lookup("logtostderr").Value.Set("true")
	if err != nil {
		glog.Fatal(err)
	}

	tagDirtySuffix := ""
	if GitTagState != "clean" {
		tagDirtySuffix = "-dirty"
	}
	treeDirtySuffix := ""
	if GitTreeState != "clean" {
		treeDirtySuffix = "-dirty"
	}
	glog.Infof("Starging K8s Virtual Device Plugin %s%s [%s%s].", Version, tagDirtySuffix, Revision, treeDirtySuffix)

	deviceFilePath := flag.Arg(0)
	if deviceFilePath == "" {
		glog.Fatalf("Device file path is required")
	}
	glog.Infof("Using device defined in %s.", deviceFilePath)

	devConfig, err := pkg.GetVirtualDeviceConfig(deviceFilePath)
	if err != nil {
		glog.Fatal(err)
	}

	vdm, err := pkg.NewVirtualDeviceManager(*devConfig)
	if err != nil {
		glog.Fatal(err)
	}

	// Respond to syscalls for termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Start grpc server
	err = vdm.Start()
	if err != nil {
		glog.Fatalf("Could not start device plugin: %v", err)
	}
	defer func() {
		err := vdm.Stop()
		if err != nil {
			glog.Error(err)
		}
	}()
	glog.Infof("Starting to serve on %s", devConfig.SocketName)

	if !*noRegister {
		// Registers with Kubelet.
		err = vdm.Register()
		if err != nil {
			glog.Fatal(err)
		}
		glog.Infof("device-plugin registered")
	}

	metricsServer := pkg.NewMetricsServer(*metricsPort, *devConfig)
	err = metricsServer.Start()
	if err != nil {
		glog.Fatal(err)
	}
	defer func() {
		err := metricsServer.Stop()
		if err != nil {
			glog.Error(err)
		}
	}()

	s := <-sigs
	glog.Infof("Received signal \"%v\", shutting down.", s)
}
