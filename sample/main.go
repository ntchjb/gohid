package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ntchjb/gohid/hid"
	"github.com/ntchjb/gohid/manager"
	"github.com/ntchjb/gohid/usb"
)

// Connect to a virtual USB echo device
// (implementation of the device located in sample folder of usbip-virtual-device),
// send strings to interrupt OUT endpoint, and receive strings from interrupt IN endpoint.
func main() {
	logger := slog.Default()
	usbCtx := usb.NewGOUSBContext()
	man := manager.NewDeviceManager(usbCtx)

	defer func() {
		if err := man.Close(); err != nil {
			logger.Error("unable to close device manager", "err", err)
		}
	}()

	// List USB HID devices
	deviceInfos, err := man.Enumerate(0, 0)
	if err != nil {
		logger.Error("unable to enumerate devices", "err", err)
		return
	}

	logger.Info("Device info", "info", deviceInfos.String())

	// Open Ledger Nano S
	hidDevice, err := man.Open(0xECC0, 0x0001, hid.DeviceConfig{
		StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
	})
	if err != nil {
		logger.Error("unable to open device")
		return
	}
	defer func() {
		if err := hidDevice.Close(); err != nil {
			logger.Error("unable to close HID device", "err", err)
		}
	}()

	if err := hidDevice.SetTarget(1, 0, 0); err != nil {
		logger.Error("unable to set target of hid device", "err", err)
		return
	}

	// Get report descriptor
	desc, err := hidDevice.GetReportDescriptor()
	if err != nil {
		logger.Error("unable to get report descriptor", "err", err)
		return
	}

	str, err := desc.String()
	if err != nil {
		logger.Error("unable to get descriptor string", "err", err)
		return
	}
	fmt.Printf("%s\n", str)
	// fmt.Printf("%v\n", desc)

	ctx := context.Background()
	for i := 0; i < 50; i++ {
		str := "Hello"
		writeData := make([]byte, 64)

		copy(writeData, []byte(str))
		n, err := hidDevice.WriteOutput(ctx, writeData)
		if err != nil {
			logger.Error("unable to write string to OUT endpoint", "err", err)
			return
		}
		logger.Info("String sent", "length", n)
	}

	for i := 0; i < 50; {
		readData := make([]byte, 64)
		n, err := hidDevice.ReadInput(ctx, readData)
		if err != nil {
			logger.Error("unable to read string from IN endpoint", "err", err)
			return
		}
		if n != 0 {
			i++
			logger.Info("String ECHO!", "length", n, "data", string(readData[:n]))
		} else {
			logger.Info("Input is empty...")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
