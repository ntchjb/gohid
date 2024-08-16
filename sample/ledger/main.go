package main

import (
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

	// Open Ledger Nano S
	device, err := man.Open(0x2C97, 0x1015)
	if err != nil {
		panic(err)
	}
	defer device.Close()

	if err := device.SetTarget(1, 0, 0); err != nil {
		panic(err)
	}

	// Get report descriptor
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
