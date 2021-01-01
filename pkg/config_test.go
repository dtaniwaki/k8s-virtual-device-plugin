package pkg

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"
)

var configFile = `
resourceName: foo
socketName: foo
count: 10
`

func TestVirtualDeviceConfig(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "k8s-virtual-device-plugin-")
	assert.NilError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(configFile))
	tmpFile.Close()

	devConfig, err := GetVirtualDeviceConfig(tmpFile.Name())
	assert.NilError(t, err)
	assert.Equal(t, *devConfig, VirtualDeviceConfig{
		ResourceName: "foo",
		SocketName:   "foo",
		Count:        10,
	})
}

func TestVirtualDeviceConfigValidate(t *testing.T) {
	{
		devConfig := VirtualDeviceConfig{
			ResourceName: "foo",
			SocketName:   "foo",
			Count:        10,
		}
		assert.NilError(t, devConfig.Validate())
	}
	{
		devConfig := VirtualDeviceConfig{
			ResourceName: "",
			SocketName:   "foo",
			Count:        10,
		}
		assert.ErrorContains(t, devConfig.Validate(), "Resource name is required.")
	}
	{
		devConfig := VirtualDeviceConfig{
			ResourceName: "foo",
			SocketName:   "",
			Count:        10,
		}
		assert.ErrorContains(t, devConfig.Validate(), "Socket name is required.")
	}
	{
		devConfig := VirtualDeviceConfig{
			ResourceName: "foo",
			SocketName:   "foo",
			Count:        0,
		}
		assert.ErrorContains(t, devConfig.Validate(), "Count must not be less than 1.")
	}
}
