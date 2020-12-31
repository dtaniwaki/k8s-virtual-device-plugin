package main

import (
	"fmt"
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
