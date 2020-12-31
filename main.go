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
	GitTreeState string = "Unknown"
)

func main() {
	noRegister := flag.Bool("no-register", false, "Run without registering to kubelet")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	dirtySuffix := ""
	if GitTreeState != "clean" {
		dirtySuffix = "-dirty"
	}
	glog.Infof("Starging K8s Virtual Device Plugin %s [%s%s].", Version, Revision, dirtySuffix)

	deviceFilePath := flag.Arg(0)
	if deviceFilePath == "" {
		glog.Fatalf("Device file path is required")
	}
	glog.Infof("Using device defined in %s.", deviceFilePath)
	vdm, err := NewVirtualDeviceManager(deviceFilePath)
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

	select {
	case s := <-sigs:
		glog.Infof("Received signal \"%v\", shutting down.", s)
		vdm.Stop()
		return
	}

}
