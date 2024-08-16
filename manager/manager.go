package manager

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/ntchjb/gohid/hid"
)

type deviceManagerImpl struct {
	goUSBCtx *gousb.Context
}

func NewDeviceManager() DeviceManager {
	return &deviceManagerImpl{}
}

func (d *deviceManagerImpl) Init() {
	d.goUSBCtx = gousb.NewContext()
}

func (d *deviceManagerImpl) Close() error {
	return d.goUSBCtx.Close()
}

func (d *deviceManagerImpl) Enumerate(vendorID gousb.ID, productID gousb.ID) (hid.DeviceInfos, error) {
	deviceInfos := hid.DeviceInfos{}
	_, err := d.goUSBCtx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if (vendorID == 0 || vendorID == desc.Vendor) && (productID == 0 || productID == desc.Product) {
			for _, config := range desc.Configs {
				for _, inf := range config.Interfaces {
					for _, setting := range inf.AltSettings {
						if setting.Class == gousb.ClassHID {
							var deviceInfo hid.DeviceInfo
							if err := deviceInfo.FromDeviceDesc(desc, config.Number, inf.Number, setting.Alternate); err != nil {
								return false
							}
							deviceInfos = append(deviceInfos, deviceInfo)
						}
					}
				}
			}
		}
		return false
	})
	if err != nil {
		return nil, fmt.Errorf("unable to open USB devices with vendorID: %d, productID: %d: %w", vendorID, productID, err)
	}

	return deviceInfos, nil
}

func (d *deviceManagerImpl) Open(vendorID, productID gousb.ID) (hid.Device, error) {
	device, err := d.goUSBCtx.OpenDeviceWithVIDPID(vendorID, productID)
	if err != nil {
		return nil, fmt.Errorf("unable to open device %v:%v: %w", vendorID, productID, err)
	}

	return hid.NewDevice(device)
}
