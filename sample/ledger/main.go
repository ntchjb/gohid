package main

import (
	"errors"
	"fmt"

	"github.com/ntchjb/gohid/manager"
)

func main() {
	man := manager.NewDeviceManager()
	man.Init()

	defer func() {
		if err := man.Close(); err != nil {
			panic(err)
		}
	}()

	deviceInfos, err := man.Enumerate(0, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("deviceInfo: %v\n", deviceInfos)

	deviceIdx := -1
	for i, deviceInfo := range deviceInfos {
		if deviceInfo.DeviceDesc.Vendor == 0x2C97 && deviceInfo.DeviceDesc.Product == 0x1015 && deviceInfo.Target[1] == 0 {
			deviceIdx = i
			break
		}
	}
	if deviceIdx < 0 {
		panic(errors.New("Ledger Nano S not found"))
	}
	device, err := man.Open(deviceInfos[deviceIdx])
	if err != nil {
		panic(err)
	}
	defer device.Close()

	desc, err := device.GetReportDescriptor()
	if err != nil {
		panic(err)
	}

	str, err := desc.String()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", str)
}
