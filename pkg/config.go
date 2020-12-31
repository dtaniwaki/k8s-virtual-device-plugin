package pkg

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

type VirtualDeviceConfig struct {
	ResourceName string `yaml:"resourceName"`
	SocketName   string `yaml:"socketName"`
	Count        int    `yaml:"count"`
}

func (vdc *VirtualDeviceConfig) Validate() error {
	if vdc.ResourceName == "" {
		return fmt.Errorf("Resource name is required.")
	}
	if vdc.SocketName == "" {
		return fmt.Errorf("Socket name is required.")
	}
	if vdc.Count < 1 {
		return fmt.Errorf("Count must not be less than 1.")
	}
	return nil
}

func GetVirtualDeviceConfig(path string) (*VirtualDeviceConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var devConfig VirtualDeviceConfig
	err = yaml.Unmarshal(raw, &devConfig)
	if err != nil {
		return nil, err
	}
	err = devConfig.Validate()
	if err != nil {
		return nil, err
	}
	glog.Infof("Device config: %v", devConfig)
	return &devConfig, nil
}
