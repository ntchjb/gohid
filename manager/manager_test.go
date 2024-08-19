package manager_test

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/gousb"
	"github.com/ntchjb/gohid/hid"
	"github.com/ntchjb/gohid/manager"
	"github.com/ntchjb/gohid/usb"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	deviceDescs = []*gousb.DeviceDesc{
		{
			Bus:                  1,
			Address:              21,
			Speed:                gousb.SpeedFull,
			Port:                 1,
			Path:                 []int{1, 2, 3},
			Spec:                 0x0200,
			Device:               0x0001,
			Vendor:               0xFF01,
			Product:              0x0001,
			Class:                gousb.ClassPerInterface,
			SubClass:             0,
			Protocol:             0,
			MaxControlPacketSize: 64,
			Configs: map[int]gousb.ConfigDesc{
				1: {
					Number:       1,
					SelfPowered:  true,
					RemoteWakeup: true,
					MaxPower:     50,
					Interfaces: []gousb.InterfaceDesc{
						{
							Number: 1,
							AltSettings: []gousb.InterfaceSetting{
								{
									Number:    1,
									Alternate: 0,
									Class:     gousb.ClassHID,
									SubClass:  gousb.Class(hid.HID_SUBCLASS_BOOT_INTERFACE),
									Protocol:  gousb.Protocol(hid.HID_PROTOCOL_MOUSE),
									Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
										0b10000001: {
											Address:       0b10000001,
											Number:        1,
											Direction:     gousb.EndpointDirectionIn,
											MaxPacketSize: 64,
											TransferType:  gousb.TransferTypeInterrupt,
											PollInterval:  10 * time.Millisecond,
										},
										0b00000001: {
											Address:       0b00000001,
											Number:        1,
											Direction:     gousb.EndpointDirectionOut,
											MaxPacketSize: 64,
											TransferType:  gousb.TransferTypeInterrupt,
											PollInterval:  10 * time.Millisecond,
										},
									},
								},
							},
						},
						{
							Number: 2,
							AltSettings: []gousb.InterfaceSetting{
								{
									Number:    2,
									Alternate: 0,
									Class:     gousb.ClassHID,
									SubClass:  gousb.Class(hid.HID_PROTOCOL_NONE),
									Protocol:  gousb.Protocol(hid.HID_PROTOCOL_MOUSE),
									Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
										0b10000010: {
											Address:       0b10000010,
											Number:        2,
											Direction:     gousb.EndpointDirectionIn,
											MaxPacketSize: 64,
											TransferType:  gousb.TransferTypeInterrupt,
											PollInterval:  10 * time.Millisecond,
										},
										0b00000010: {
											Address:       0b00000010,
											Number:        2,
											Direction:     gousb.EndpointDirectionOut,
											MaxPacketSize: 64,
											TransferType:  gousb.TransferTypeInterrupt,
											PollInterval:  10 * time.Millisecond,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Bus:                  2,
			Address:              20,
			Speed:                gousb.SpeedHigh,
			Port:                 2,
			Path:                 []int{1, 2, 4},
			Spec:                 0x0200,
			Device:               0x0001,
			Vendor:               0xFF02,
			Product:              0x0001,
			Class:                gousb.ClassPerInterface,
			SubClass:             0,
			Protocol:             0,
			MaxControlPacketSize: 64,
			Configs: map[int]gousb.ConfigDesc{
				1: {
					Number:       1,
					SelfPowered:  true,
					RemoteWakeup: true,
					MaxPower:     50,
					Interfaces: []gousb.InterfaceDesc{
						{
							Number: 1,
							AltSettings: []gousb.InterfaceSetting{
								{
									Number:    1,
									Alternate: 0,
									Class:     gousb.ClassAudio,
									SubClass:  1,
									Protocol:  0,
									Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
										0b10000001: {
											Address:       0b10000001,
											Number:        1,
											Direction:     gousb.EndpointDirectionIn,
											MaxPacketSize: 64,
											TransferType:  gousb.TransferTypeInterrupt,
											PollInterval:  10 * time.Millisecond,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
)

func TestDeviceManager_Enumerate(t *testing.T) {
	errBadAccess := errors.New("bad access")
	type fields struct {
		usbCtx func(ctrl *gomock.Controller) usb.Context
	}
	type args struct {
		vendorID  gousb.ID
		productID gousb.ID
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		res    func(t *testing.T) hid.DeviceInfos
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				usbCtx: func(ctrl *gomock.Controller) usb.Context {
					usbCtx := usb.NewMockContext(ctrl)
					usbCtx.EXPECT().IterateDevices(gomock.Any()).DoAndReturn(func(reader func(desc *gousb.DeviceDesc)) error {
						reader(deviceDescs[0])
						reader(deviceDescs[1])

						return nil
					})

					return usbCtx
				},
			},
			args: args{
				vendorID:  0,
				productID: 0,
			},
			res: func(t *testing.T) hid.DeviceInfos {
				var device1 hid.DeviceInfo
				var device2 hid.DeviceInfo

				err := device1.FromDeviceDesc(deviceDescs[0], 1, 1, 0)
				assert.NoError(t, err)
				err = device2.FromDeviceDesc(deviceDescs[0], 1, 2, 0)
				assert.NoError(t, err)

				return hid.DeviceInfos{
					device1, device2,
				}
			},
			err: nil,
		},
		{
			name: "Error IterateDevices",
			fields: fields{
				usbCtx: func(ctrl *gomock.Controller) usb.Context {
					usbCtx := usb.NewMockContext(ctrl)
					usbCtx.EXPECT().IterateDevices(gomock.Any()).DoAndReturn(func(reader func(desc *gousb.DeviceDesc)) error {
						return errBadAccess
					})

					return usbCtx
				},
			},
			args: args{
				vendorID:  0,
				productID: 0,
			},
			res: func(t *testing.T) hid.DeviceInfos {
				return nil
			},
			err: errBadAccess,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			usbCtx := test.fields.usbCtx(ctrl)
			logger := slog.Default()
			man := manager.NewDeviceManager(usbCtx, logger)

			deviceInfos, err := man.Enumerate(test.args.vendorID, test.args.productID)
			expected := test.res(t)

			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, expected, deviceInfos)
		})
	}
}

func TestDeviceManager_Open(t *testing.T) {
	errBadAccess := errors.New("bad access")
	type fields struct {
		usbCtx func(ctrl *gomock.Controller, vendorID, productID gousb.ID) (usb.Context, usb.Device)
	}
	type args struct {
		vendorID  gousb.ID
		productID gousb.ID
		config    hid.DeviceConfig
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		res    func(t *testing.T, mockDevice *usb.MockDevice) hid.Device
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				usbCtx: func(ctrl *gomock.Controller, vendorID, productID gousb.ID) (usb.Context, usb.Device) {
					usbCtx := usb.NewMockContext(ctrl)
					usbDevice := usb.NewMockDevice(ctrl)
					usbCtx.EXPECT().OpenDevice(vendorID, productID).Return(usbDevice, nil)

					return usbCtx, usbDevice
				},
			},
			args: args{
				vendorID:  0xFF11,
				productID: 0x0001,
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			res: func(t *testing.T, mockDevice *usb.MockDevice) hid.Device {
				logger := slog.Default()
				device, err := hid.NewDevice(mockDevice, hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				}, logger)
				assert.NoError(t, err)

				return device
			},
			err: nil,
		},
		{
			name: "Error_BadAccess",
			fields: fields{
				usbCtx: func(ctrl *gomock.Controller, vendorID, productID gousb.ID) (usb.Context, usb.Device) {
					usbCtx := usb.NewMockContext(ctrl)
					usbCtx.EXPECT().OpenDevice(vendorID, productID).Return(nil, errBadAccess)

					return usbCtx, nil
				},
			},
			args: args{
				vendorID:  0xFF11,
				productID: 0x0001,
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			res: func(t *testing.T, mockDevice *usb.MockDevice) hid.Device {
				return nil
			},
			err: errBadAccess,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			usbCtx, usbDevice := test.fields.usbCtx(ctrl, test.args.vendorID, test.args.productID)
			logger := slog.Default()
			man := manager.NewDeviceManager(usbCtx, logger)
			var expectedHIDDevice hid.Device
			if usbDevice != nil {
				expectedHIDDevice = test.res(t, usbDevice.(*usb.MockDevice))
			}

			hidDevice, err := man.Open(test.args.vendorID, test.args.productID, test.args.config)

			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, expectedHIDDevice, hidDevice)
		})
	}
}
