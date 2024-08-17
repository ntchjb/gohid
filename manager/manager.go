package manager

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/ntchjb/gohid/hid"
	"github.com/ntchjb/gohid/usb"
)

type DeviceManager interface {
	Close() error
	Enumerate(vendorID gousb.ID, productID gousb.ID) (hid.DeviceInfos, error)
	Open(vendorID, productID gousb.ID, config hid.DeviceConfig) (hid.Device, error)
}

type deviceManagerImpl struct {
	goUSBCtx usb.Context
}

func NewDeviceManager(ctx usb.Context) DeviceManager {
	return &deviceManagerImpl{
		goUSBCtx: ctx,
	}
}

func (d *deviceManagerImpl) Close() error {
	return d.goUSBCtx.Close()
}

func (d *deviceManagerImpl) Enumerate(vendorID gousb.ID, productID gousb.ID) (hid.DeviceInfos, error) {
	deviceInfos := hid.DeviceInfos{}
	if err := d.goUSBCtx.IterateDevices(func(desc *gousb.DeviceDesc) {
		if (vendorID == 0 || vendorID == desc.Vendor) && (productID == 0 || productID == desc.Product) {
			for _, config := range desc.Configs {
				for _, inf := range config.Interfaces {
					for _, setting := range inf.AltSettings {
						if setting.Class == gousb.ClassHID {
							var deviceInfo hid.DeviceInfo
							if err := deviceInfo.FromDeviceDesc(desc, config.Number, inf.Number, setting.Alternate); err != nil {
								return
							}
							deviceInfos = append(deviceInfos, deviceInfo)
						}
					}
				}
			}
		}
	}); err != nil {
		return nil, fmt.Errorf("unable to open USB devices with vendorID: %d, productID: %d: %w", vendorID, productID, err)
	}

	return deviceInfos, nil
}

func (d *deviceManagerImpl) Open(vendorID, productID gousb.ID, config hid.DeviceConfig) (hid.Device, error) {
	device, err := d.goUSBCtx.OpenDevice(vendorID, productID)
	if err != nil {
		return nil, fmt.Errorf("unable to open device %v:%v: %w", vendorID, productID, err)
	}

	return hid.NewDevice(device, config)
}
