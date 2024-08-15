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

func (d *deviceManagerImpl) Enumerate(vendorID uint16, productID uint16) (hid.DeviceInfos, error) {
	deviceInfos := hid.DeviceInfos{}
	_, err := d.goUSBCtx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if (vendorID == 0 || vendorID == uint16(desc.Vendor)) && (productID == 0 || productID == uint16(desc.Product)) {
			for _, config := range desc.Configs {
				for _, inf := range config.Interfaces {
					for _, setting := range inf.AltSettings {
						if setting.Class == gousb.ClassHID {
							device := hid.DeviceInfo{
								DeviceDesc: *desc,
								ConfigDesc: config,
								Target:     [3]int{config.Number, inf.Number, setting.Alternate},
								Protocol:   hid.HIDProtocol(setting.Protocol),
								Subclass:   hid.SubClass(setting.SubClass),
							}
							for _, endpoint := range setting.Endpoints {
								device.Endpoints = append(device.Endpoints, hid.DeviceEndpoint{
									Number:        endpoint.Number,
									Address:       uint8(endpoint.Address),
									Direction:     bool(endpoint.Direction),
									MaxPacketSize: endpoint.MaxPacketSize,
									PollInterval:  endpoint.PollInterval,
									TransferType:  hid.EndpointTransferType(endpoint.TransferType),
								})
							}
							deviceInfos = append(deviceInfos, device)
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

func (d *deviceManagerImpl) Open(deviceInfo hid.DeviceInfo) (hid.DeviceAccessor, error) {
	device, err := d.goUSBCtx.OpenDeviceWithVIDPID(deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product)
	if err != nil {
		return nil, fmt.Errorf("unable to open device %v:%v: %w", deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product, err)
	}

	return hid.NewDevice(device, deviceInfo)
}
