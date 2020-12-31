package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

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
	flag.Lookup("logtostderr").Value.Set("true")

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

	devConfig, err := GetVirtualDeviceConfig(deviceFilePath)
	if err != nil {
		glog.Fatal(err)
	}

	vdm, err := NewVirtualDeviceManager(*devConfig)
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
	glog.Infof("Starting to serve on %s", vdm.socketName)

	if !*noRegister {
		// Registers with Kubelet.
		err = vdm.Register()
		if err != nil {
			glog.Fatal(err)
		}
		glog.Infof("device-plugin registered")
	}

	metricsServer := NewMetricsServer(*metricsPort, *devConfig)
	metricsServer.Start()

	select {
	case s := <-sigs:
		glog.Infof("Received signal \"%v\", shutting down.", s)
		vdm.Stop()
		metricsServer.Stop()
		return
	}

}
