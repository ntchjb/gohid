package hid_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/gousb"
	"github.com/ntchjb/gohid/hid"
	"github.com/ntchjb/gohid/usb"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ntchjb/usbip-virtual-device/usb/protocol/descriptor"
	hidprotocol "github.com/ntchjb/usbip-virtual-device/usb/protocol/hid"
	hidreport "github.com/ntchjb/usbip-virtual-device/usb/protocol/hid/report"
)

type mocks struct {
	device *usb.MockDevice
	config *usb.MockConfig
	inf    *usb.MockInterface
	epIn   *usb.MockInEndpoint
	epOut  *usb.MockOutEndpoint
	reader *usb.MockStreamReader
	writer *usb.MockStreamWriter
}

var (
	deviceDesc = &gousb.DeviceDesc{
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
					{
						Number: 3,
						AltSettings: []gousb.InterfaceSetting{
							{
								Number:    3,
								Alternate: 0,
								Class:     gousb.ClassHID,
								SubClass:  gousb.Class(hid.HID_PROTOCOL_NONE),
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
								},
							},
						},
					},
				},
			},
		},
	}

	createSetupTargetMocks = func(ctrl *gomock.Controller, confNum, infNum, altNum, epIn, epOut int) mocks {
		mockUSBDevice := usb.NewMockDevice(ctrl)
		mockUSBConfig := usb.NewMockConfig(ctrl)
		mockUSBInf := usb.NewMockInterface(ctrl)
		mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
		mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
		mockUSBStreamReader := usb.NewMockStreamReader(ctrl)
		mockUSBStreamWriter := usb.NewMockStreamWriter(ctrl)
		mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
		mockUSBDevice.EXPECT().Config(confNum).Return(mockUSBConfig, nil)
		mockUSBConfig.EXPECT().Interface(infNum, altNum).Return(mockUSBInf, nil)
		mockUSBInf.EXPECT().InEndpoint(epIn).Return(mockUSBInEndpoint, nil)
		mockUSBInf.EXPECT().OutEndpoint(epOut).Return(mockUSBOutEndpoint, nil).AnyTimes()
		mockUSBInEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamReader, nil)
		mockUSBOutEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamWriter, nil).AnyTimes()

		return mocks{
			device: mockUSBDevice,
			config: mockUSBConfig,
			inf:    mockUSBInf,
			epIn:   mockUSBInEndpoint,
			epOut:  mockUSBOutEndpoint,
			reader: mockUSBStreamReader,
			writer: mockUSBStreamWriter,
		}
	}

	config = hid.DeviceConfig{
		StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
	}
)

func TestDevice_NewDevice(t *testing.T) {
	type fields struct {
		device func(ctrl *gomock.Controller) usb.Device
		config hid.DeviceConfig
	}

	tests := []struct {
		name   string
		fields fields
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				device: func(ctrl *gomock.Controller) usb.Device {
					mockUSBDevice := usb.NewMockDevice(ctrl)

					return mockUSBDevice
				},
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			err: nil,
		},
		{
			name: "Error_DeviceIsNil",
			fields: fields{
				device: func(ctrl *gomock.Controller) usb.Device {
					return nil
				},
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			err: hid.ErrDeviceIsNil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usbDevice := test.fields.device(ctrl)
			logger := slog.Default()
			_, err := hid.NewDevice(usbDevice, test.fields.config, logger)

			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestDevice_SetAutoDetach(t *testing.T) {
	errUnknown := errors.New("unknown")
	type fields struct {
		device func(ctrl *gomock.Controller) usb.Device
		config hid.DeviceConfig
	}
	type args struct {
		autoDetach bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				device: func(ctrl *gomock.Controller) usb.Device {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBDevice.EXPECT().SetAutoDetach(true).Return(nil)

					return mockUSBDevice
				},
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			args: args{
				autoDetach: true,
			},
			err: nil,
		},
		{
			name: "Success_SetFalse",
			fields: fields{
				device: func(ctrl *gomock.Controller) usb.Device {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBDevice.EXPECT().SetAutoDetach(false).Return(nil)

					return mockUSBDevice
				},
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			args: args{
				autoDetach: false,
			},
			err: nil,
		},
		{
			name: "Error",
			fields: fields{
				device: func(ctrl *gomock.Controller) usb.Device {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBDevice.EXPECT().SetAutoDetach(true).Return(errUnknown)
					mockUSBDevice.EXPECT().Descriptor().Return(&gousb.DeviceDesc{
						Vendor:  0x1234,
						Product: 0x2345,
					})

					return mockUSBDevice
				},
				config: hid.DeviceConfig{
					StreamLaneCount: hid.DEFAULT_ENDPOINT_STREAM_COUNT,
				},
			},
			args: args{
				autoDetach: true,
			},
			err: errUnknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usbDevice := test.fields.device(ctrl)
			logger := slog.Default()
			hidDevice, err := hid.NewDevice(usbDevice, test.fields.config, logger)

			assert.NoError(t, err)

			err = hidDevice.SetAutoDetach(test.args.autoDetach)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestDevice_SetTarget(t *testing.T) {
	errBadAccess := errors.New("bad access")
	errMemoryIssue := errors.New("err memory issue")

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		confNumber int
		infNumber  int
		altNumber  int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBStreamReader := usb.NewMockStreamReader(ctrl)
					mockUSBStreamWriter := usb.NewMockStreamWriter(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil)
					mockUSBInf.EXPECT().InEndpoint(1).Return(mockUSBInEndpoint, nil)
					mockUSBInf.EXPECT().OutEndpoint(1).Return(mockUSBOutEndpoint, nil)
					mockUSBInEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamReader, nil)
					mockUSBOutEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamWriter, nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
						epOut:  mockUSBOutEndpoint,
						reader: mockUSBStreamReader,
						writer: mockUSBStreamWriter,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: nil,
		},
		{
			name: "Error_CannotClaimConfiguration",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(nil, errBadAccess)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: errBadAccess,
		},
		{
			name: "Error_CannotClaimInterface",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(nil, errBadAccess)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: errBadAccess,
		},
		{
			name: "Error_CannotClaimEndpointIn",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil)
					mockUSBInf.EXPECT().InEndpoint(1).Return(nil, errBadAccess)
					// iterate endpoints on map, order of endpoints is not guaranteed
					mockUSBInf.EXPECT().OutEndpoint(1).Return(mockUSBOutEndpoint, nil).AnyTimes()
					mockUSBInf.EXPECT().Close().Return(nil)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: errBadAccess,
		},
		{
			name: "Error_CannotClaimEndpointOut",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil)
					// iterate endpoints on map, order of endpoints is not guaranteed
					mockUSBInf.EXPECT().InEndpoint(1).Return(mockUSBInEndpoint, nil).AnyTimes()
					mockUSBInf.EXPECT().OutEndpoint(1).Return(nil, errBadAccess)
					mockUSBInf.EXPECT().Close().Return(nil)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: errBadAccess,
		},
		{
			name: "Error_EndpointInNotFound",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(2, 0).Return(mockUSBInf, nil)
					mockUSBInf.EXPECT().OutEndpoint(2).Return(mockUSBOutEndpoint, nil)
					mockUSBInf.EXPECT().Close().Return(nil)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
						epOut:  mockUSBOutEndpoint,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  2,
				altNumber:  0,
			},
			err: hid.ErrEndpointInNotFound,
		},
		{
			name: "Error_CannotCreateStreamReader",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil)
					mockUSBInf.EXPECT().InEndpoint(1).Return(mockUSBInEndpoint, nil)
					mockUSBInf.EXPECT().OutEndpoint(1).Return(mockUSBOutEndpoint, nil)
					mockUSBInEndpoint.EXPECT().Descriptor().Return(deviceDesc.Configs[1].Interfaces[0].AltSettings[0].Endpoints[0b10000001])
					mockUSBInEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(nil, errMemoryIssue)
					mockUSBInf.EXPECT().Close().Return(nil)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
						epOut:  mockUSBOutEndpoint,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: errMemoryIssue,
		},
		{
			name: "Error_CannotCreateStreamWriter",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBStreamReader := usb.NewMockStreamReader(ctrl)
					mockUSBStreamWriter := usb.NewMockStreamWriter(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil)
					mockUSBInf.EXPECT().InEndpoint(1).Return(mockUSBInEndpoint, nil)
					mockUSBInf.EXPECT().OutEndpoint(1).Return(mockUSBOutEndpoint, nil)
					mockUSBInEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamReader, nil)
					mockUSBOutEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(nil, errMemoryIssue)
					mockUSBOutEndpoint.EXPECT().Descriptor().Return(deviceDesc.Configs[1].Interfaces[0].AltSettings[0].Endpoints[0b00000001])
					mockUSBInf.EXPECT().Close().Return(nil)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
						epOut:  mockUSBOutEndpoint,
						reader: mockUSBStreamReader,
						writer: mockUSBStreamWriter,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: errMemoryIssue,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mocks := test.fields.mocks(ctrl)
			logger := slog.Default()

			hidDevice, err := hid.NewDevice(mocks.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(test.args.confNumber, test.args.infNumber, test.args.altNumber)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestDevice_SetTarget_Multiple(t *testing.T) {
	errCloseFailed := errors.New("error closing resource")

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		confNumber int
		infNumber  int
		altNumber  int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBStreamReader := usb.NewMockStreamReader(ctrl)
					mockUSBStreamWriter := usb.NewMockStreamWriter(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc).Times(2)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil).Times(2)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil).Times(2)
					mockUSBInf.EXPECT().InEndpoint(1).Return(mockUSBInEndpoint, nil).Times(2)
					mockUSBInf.EXPECT().OutEndpoint(1).Return(mockUSBOutEndpoint, nil).Times(2)
					mockUSBInEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamReader, nil).Times(2)
					mockUSBOutEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamWriter, nil).Times(2)
					mockUSBStreamReader.EXPECT().Close().Return(nil)
					mockUSBStreamWriter.EXPECT().Close().Return(nil)
					mockUSBInf.EXPECT().Close().Return(nil)
					mockUSBConfig.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
						epOut:  mockUSBOutEndpoint,
						reader: mockUSBStreamReader,
						writer: mockUSBStreamWriter,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: nil,
		},
		{
			name: "Error Closing Existing Resources",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBConfig := usb.NewMockConfig(ctrl)
					mockUSBInf := usb.NewMockInterface(ctrl)
					mockUSBInEndpoint := usb.NewMockInEndpoint(ctrl)
					mockUSBOutEndpoint := usb.NewMockOutEndpoint(ctrl)
					mockUSBStreamReader := usb.NewMockStreamReader(ctrl)
					mockUSBStreamWriter := usb.NewMockStreamWriter(ctrl)
					mockUSBDevice.EXPECT().Descriptor().Return(deviceDesc).Times(2)
					mockUSBDevice.EXPECT().Config(1).Return(mockUSBConfig, nil).Times(2)
					mockUSBConfig.EXPECT().Interface(1, 0).Return(mockUSBInf, nil).Times(2)
					mockUSBInf.EXPECT().InEndpoint(1).Return(mockUSBInEndpoint, nil).Times(2)
					mockUSBInf.EXPECT().OutEndpoint(1).Return(mockUSBOutEndpoint, nil).Times(2)
					mockUSBInEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamReader, nil).Times(2)
					mockUSBOutEndpoint.EXPECT().NewStream(hid.DEFAULT_ENDPOINT_STREAM_COUNT).Return(mockUSBStreamWriter, nil).Times(2)
					mockUSBStreamReader.EXPECT().Close().Return(errCloseFailed)
					mockUSBStreamWriter.EXPECT().Close().Return(errCloseFailed)
					mockUSBInf.EXPECT().Close().Return(errCloseFailed)
					mockUSBConfig.EXPECT().Close().Return(errCloseFailed)

					return mocks{
						device: mockUSBDevice,
						config: mockUSBConfig,
						inf:    mockUSBInf,
						epIn:   mockUSBInEndpoint,
						epOut:  mockUSBOutEndpoint,
						reader: mockUSBStreamReader,
						writer: mockUSBStreamWriter,
					}
				},
				config: config,
			},
			args: args{
				confNumber: 1,
				infNumber:  1,
				altNumber:  0,
			},
			err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mocks := test.fields.mocks(ctrl)
			logger := slog.Default()

			hidDevice, err := hid.NewDevice(mocks.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(test.args.confNumber, test.args.infNumber, test.args.altNumber)
			assert.ErrorIs(t, err, test.err)
			err = hidDevice.SetTarget(test.args.confNumber, test.args.infNumber, test.args.altNumber)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestDevice_Close(t *testing.T) {
	errCloseReader := errors.New("error close reader")
	errCloseWriter := errors.New("error close writer")
	errCloseInf := errors.New("error close interface")
	errCloseConfiguration := errors.New("error close configuration")
	errCloseDevice := errors.New("error close device")
	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}

	tests := []struct {
		name   string
		preRun func(t *testing.T, hidDevice hid.Device)
		fields fields
		errs   []error
	}{
		{
			name: "Success_CloseAllResources",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.reader.EXPECT().Close().Return(nil)
					mocks.writer.EXPECT().Close().Return(nil)
					mocks.inf.EXPECT().Close().Return(nil)
					mocks.config.EXPECT().Close().Return(nil)
					mocks.device.EXPECT().Close().Return(nil)

					return mocks
				},
				config: config,
			},
			preRun: func(t *testing.T, hidDevice hid.Device) {
				err := hidDevice.SetTarget(1, 1, 0)
				assert.ErrorIs(t, err, nil)
			},
			errs: nil,
		},
		{
			name: "Success_NoSetupTarget_CloseOnlyDevice",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mockUSBDevice := usb.NewMockDevice(ctrl)
					mockUSBDevice.EXPECT().Close().Return(nil)

					return mocks{
						device: mockUSBDevice,
					}
				},
				config: config,
			},
			preRun: func(t *testing.T, hidDevice hid.Device) {},
			errs:   nil,
		},
		{
			name: "Error_ClosingResources",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.reader.EXPECT().Close().Return(errCloseReader)
					mocks.writer.EXPECT().Close().Return(errCloseWriter)
					mocks.inf.EXPECT().Close().Return(errCloseInf)
					mocks.config.EXPECT().Close().Return(errCloseConfiguration)
					mocks.device.EXPECT().Close().Return(errCloseDevice)

					return mocks
				},
				config: config,
			},
			preRun: func(t *testing.T, hidDevice hid.Device) {
				err := hidDevice.SetTarget(1, 1, 0)
				assert.ErrorIs(t, err, nil)
			},
			errs: []error{
				errCloseReader,
				errCloseWriter,
				errCloseInf,
				errCloseConfiguration,
				errCloseDevice,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mocks := test.fields.mocks(ctrl)
			logger := slog.Default()

			hidDevice, err := hid.NewDevice(mocks.device, test.fields.config, logger)
			assert.NoError(t, err)

			test.preRun(t, hidDevice)

			err = hidDevice.Close()
			for _, targetErr := range test.errs {
				assert.ErrorIs(t, err, targetErr)
			}
		})
	}
}

func TestDevice_WriteOutput(t *testing.T) {
	errInterrupt := errors.New("interrupt transfer error")
	ctx := context.Background()

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		ctx    context.Context
		data   []byte
		infNum int
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		byteWritten int
		err         error
	}{
		{
			name: "Success_WithoutReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.writer.EXPECT().WriteContext(ctx, []byte{
						0x01, 0x02, 0x03, 0x04, 0x05,
					}).Return(5, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Success_WithReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.writer.EXPECT().WriteContext(ctx, []byte{
						0x01, 0x01, 0x02, 0x03, 0x04, 0x05,
					}).Return(6, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Success_WriteToControlEndpoint",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 3, 0, 1, 1)

					mocks.device.EXPECT().Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0200), uint16(0x0003), []byte{0x01, 0x02, 0x03, 0x04, 0x05}).Return(5, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 3,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Error_WriteContext",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.writer.EXPECT().WriteContext(ctx, []byte{
						0x01, 0x02, 0x03, 0x04, 0x05,
					}).Return(0, errInterrupt)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 0,
			err:         errInterrupt,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.args.infNum, 0)
			assert.NoError(t, err)

			byteWritten, err := hidDevice.WriteOutput(test.args.ctx, test.args.data)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.byteWritten, byteWritten)
		})
	}
}

func TestDevice_ReadInput(t *testing.T) {
	errInterrupt := errors.New("interrupt transfer error")
	ctx := context.Background()

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		ctx    context.Context
		data   []byte
		infNum int
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		byteRead int
		err      error
	}{
		{
			name: "Success",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)
					mocks.reader.EXPECT().
						ReadContext(ctx, make([]byte, 6)).
						DoAndReturn(func(ctx context.Context, data []byte) (int, error) {
							readData := []byte{0x01, 0x01, 0x02, 0x03, 0x04, 0x05}
							copy(data, readData)

							return len(readData), nil
						})

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   make([]byte, 6),
				infNum: 1,
			},
			byteRead: 6,
			err:      nil,
		},
		{
			name: "Success_ZeroLengthData",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)
					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   nil,
				infNum: 1,
			},
			byteRead: 0,
			err:      nil,
		},
		{
			name: "Error_ReadContext",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)
					mocks.reader.EXPECT().
						ReadContext(ctx, make([]byte, 6)).
						DoAndReturn(func(ctx context.Context, data []byte) (int, error) {
							return 0, errInterrupt
						})

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   make([]byte, 6),
				infNum: 1,
			},
			byteRead: 0,
			err:      errInterrupt,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.args.infNum, 0)
			assert.NoError(t, err)

			byteWritten, err := hidDevice.ReadInput(test.args.ctx, test.args.data)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.byteRead, byteWritten)
		})
	}
}

func TestDevice_SendFeatureReport(t *testing.T) {
	errControl := errors.New("control transfer error")
	ctx := context.Background()

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		ctx    context.Context
		data   []byte
		infNum int
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		byteWritten int
		err         error
	}{
		{
			name: "Success_WithReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0301), uint16(1), []byte{
							0x01, 0x01, 0x02, 0x03, 0x04, 0x05,
						}).
						Return(6, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Success_WithoutReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0300), uint16(1), []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						}).
						Return(5, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Error_EmptyReportData",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   nil,
				infNum: 1,
			},
			byteWritten: 0,
			err:         hid.ErrEmptyData,
		},
		{
			name: "Error_SendRequestViaControl",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0300), uint16(1), []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						}).
						Return(0, errControl)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 0,
			err:         errControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.args.infNum, 0)
			assert.NoError(t, err)

			byteWritten, err := hidDevice.SendFeatureReport(test.args.data)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.byteWritten, byteWritten)
		})
	}
}

func TestDevice_GetFeatureReport(t *testing.T) {
	errControl := errors.New("control transfer error")
	ctx := context.Background()

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		ctx    context.Context
		data   []byte
		infNum int
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		byteRead int
		err      error
	}{
		{
			name: "Success_WithReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1010_0001), uint8(0x01), uint16(0x0301), uint16(1), []byte{
							0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
						}).DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
						readData := []byte{0x01, 0x01, 0x02, 0x03, 0x04, 0x05}
						copy(data, readData)

						return len(readData), nil
					})

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
				infNum: 1,
			},
			byteRead: 6,
			err:      nil,
		},
		{
			name: "Success_WithoutReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1010_0001), uint8(0x01), uint16(0x0300), uint16(1), []byte{
							0x00, 0x00, 0x00, 0x00, 0x00,
						}).DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
						readData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
						copy(data, readData)

						return len(readData), nil
					})

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				infNum: 1,
			},
			byteRead: 6,
			err:      nil,
		},
		{
			name: "Error_EmptyReportData",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   nil,
				infNum: 1,
			},
			byteRead: 0,
			err:      hid.ErrEmptyData,
		},
		{
			name: "Error_SendRequestViaControl",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1010_0001), uint8(0x01), uint16(0x0301), uint16(1), []byte{
							0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
						}).
						Return(0, errControl)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
				infNum: 1,
			},
			byteRead: 0,
			err:      errControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.args.infNum, 0)
			assert.NoError(t, err)

			byteWritten, err := hidDevice.GetFeatureReport(test.args.data)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.byteRead, byteWritten)
		})
	}
}

func TestDevice_SendOutputReport(t *testing.T) {
	errControl := errors.New("control transfer error")
	ctx := context.Background()

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		ctx    context.Context
		data   []byte
		infNum int
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		byteWritten int
		err         error
	}{
		{
			name: "Success_WithReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0201), uint16(1), []byte{
							0x01, 0x01, 0x02, 0x03, 0x04, 0x05,
						}).
						Return(6, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Success_WithoutReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0200), uint16(1), []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						}).
						Return(5, nil)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 6,
			err:         nil,
		},
		{
			name: "Error_EmptyReportData",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   nil,
				infNum: 1,
			},
			byteWritten: 0,
			err:         hid.ErrEmptyData,
		},
		{
			name: "Error_SendRequestViaControl",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b0010_0001), uint8(0x09), uint16(0x0200), uint16(1), []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						}).
						Return(0, errControl)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				infNum: 1,
			},
			byteWritten: 0,
			err:         errControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.args.infNum, 0)
			assert.NoError(t, err)

			byteWritten, err := hidDevice.SendOutputReport(test.args.data)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.byteWritten, byteWritten)
		})
	}
}

func TestDevice_GetInputReport(t *testing.T) {
	errControl := errors.New("control transfer error")
	ctx := context.Background()

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}
	type args struct {
		ctx    context.Context
		data   []byte
		infNum int
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		byteRead int
		err      error
	}{
		{
			name: "Success_WithReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1010_0001), uint8(0x01), uint16(0x0101), uint16(1), []byte{
							0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
						}).DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
						readData := []byte{0x01, 0x01, 0x02, 0x03, 0x04, 0x05}
						copy(data, readData)

						return len(readData), nil
					})

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
				infNum: 1,
			},
			byteRead: 6,
			err:      nil,
		},
		{
			name: "Success_WithoutReportID",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1010_0001), uint8(0x01), uint16(0x0100), uint16(1), []byte{
							0x00, 0x00, 0x00, 0x00, 0x00,
						}).DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
						readData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
						copy(data, readData)

						return len(readData), nil
					})

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				infNum: 1,
			},
			byteRead: 6,
			err:      nil,
		},
		{
			name: "Error_EmptyReportData",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   nil,
				infNum: 1,
			},
			byteRead: 0,
			err:      hid.ErrEmptyData,
		},
		{
			name: "Error_SendRequestViaControl",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1010_0001), uint8(0x01), uint16(0x0101), uint16(1), []byte{
							0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
						}).
						Return(0, errControl)

					return mocks
				},
				config: config,
			},
			args: args{
				ctx:    ctx,
				data:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
				infNum: 1,
			},
			byteRead: 0,
			err:      errControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.args.infNum, 0)
			assert.NoError(t, err)

			byteWritten, err := hidDevice.GetInputReport(test.args.data)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.byteRead, byteWritten)
		})
	}
}

func TestDevice_GetMetadata(t *testing.T) {
	manufacturer := "Awesome factory Co.Ltd."
	productName := "Awesome product"
	serialNumber := "1ABF4968DE"
	errStringDescNotFound := errors.New("string desc not found")
	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
		infNum int
	}

	tests := []struct {
		name         string
		fields       fields
		manufacturer string
		productName  string
		serialNumber string
		err          error
	}{
		{
			name: "Success",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().Manufacturer().Return(manufacturer, nil)
					mocks.device.EXPECT().Product().Return(productName, nil)
					mocks.device.EXPECT().SerialNumber().Return(serialNumber, nil)

					return mocks
				},
				config: config,
				infNum: 1,
			},
			manufacturer: manufacturer,
			productName:  productName,
			serialNumber: serialNumber,
			err:          nil,
		},
		{
			name: "Error_StringDescError",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().Manufacturer().Return("", errStringDescNotFound)
					mocks.device.EXPECT().Product().Return("", errStringDescNotFound)
					mocks.device.EXPECT().SerialNumber().Return("", errStringDescNotFound)

					return mocks
				},
				config: config,
				infNum: 1,
			},
			err: errStringDescNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, test.fields.infNum, 0)
			assert.NoError(t, err)

			manu, err1 := hidDevice.GetManufacturer()
			prod, err2 := hidDevice.GetProduct()
			seri, err3 := hidDevice.GetSerialNumber()

			assert.Equal(t, test.manufacturer, manu)
			assert.Equal(t, test.productName, prod)
			assert.Equal(t, test.serialNumber, seri)

			assert.ErrorIs(t, err1, test.err)
			assert.ErrorIs(t, err2, test.err)
			assert.ErrorIs(t, err3, test.err)
		})
	}
}

func TestDevice_GetReportDescriptor(t *testing.T) {
	errControl := errors.New("control transfer error")

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}

	tests := []struct {
		name   string
		fields fields
		desc   hidreport.HIDReportDescriptor
		err    error
	}{
		{
			name: "Success",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2200), uint16(1), make([]byte, hid.HID_MAX_REPORT_SIZE)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							readData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
							copy(data, readData)

							return len(readData), nil
						})

					return mocks
				},
				config: config,
			},
			desc: hidreport.HIDReportDescriptor{0x01, 0x02, 0x03, 0x04, 0x05},
			err:  nil,
		},
		{
			name: "Error_ControlTransferError",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2200), uint16(1), make([]byte, hid.HID_MAX_REPORT_SIZE)).
						Return(0, errControl)
					return mocks
				},
				config: config,
			},
			desc: nil,
			err:  errControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, 1, 0)
			assert.NoError(t, err)

			desc, err := hidDevice.GetReportDescriptor()
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.desc, desc)
		})
	}
}

func TestDevice_GetHIDDescriptor(t *testing.T) {
	errControl := errors.New("control transfer error")

	type fields struct {
		mocks  func(ctrl *gomock.Controller) mocks
		config hid.DeviceConfig
	}

	tests := []struct {
		name   string
		fields fields
		desc   hidprotocol.HIDDescriptor
		err    error
	}{
		{
			name: "Success_SingleDescriptor",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2100), uint16(1), make([]byte, hidprotocol.HID_DESCRIPTOR_LENGTH)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							readData := []byte{
								0x09,
								0x21,
								0x10, 0x01,
								0x01,
								0x01,
								0x22,
								0x3F, 0x00,
							}
							copy(data, readData)

							return len(readData), nil
						})

					return mocks
				},
				config: config,
			},
			desc: hidprotocol.HIDDescriptor{
				BLength:              hidprotocol.HID_DESCRIPTOR_LENGTH,
				BDescriptorType:      descriptor.DESCRIPTOR_TYPE_HID,
				BCDHID:               0x0110,
				BCountryCode:         0x01,
				BNumDescriptors:      0x01,
				BClassDescriptorType: 0x22,
				WDescriptorLength:    0x003F,
			},
			err: nil,
		},
		{
			name: "Success_MultipleDescriptor",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)
					readData := []byte{
						0x0c,
						0x21,
						0x10, 0x01,
						0x01,
						0x03,
						0x22,
						0x3F, 0x00,
						0x22,
						0x41, 0x00,
						0x22,
						0x42, 0x00,
					}

					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2100), uint16(1), make([]byte, hidprotocol.HID_DESCRIPTOR_LENGTH)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							copy(data, readData)

							return len(data), nil
						})
					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2100), uint16(1), make([]byte, hidprotocol.HID_DESCRIPTOR_LENGTH+6)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							copy(data, readData)

							return len(data), nil
						})

					return mocks
				},
				config: config,
			},
			desc: hidprotocol.HIDDescriptor{
				BLength:              hidprotocol.HID_DESCRIPTOR_LENGTH + 3,
				BDescriptorType:      descriptor.DESCRIPTOR_TYPE_HID,
				BCDHID:               0x0110,
				BCountryCode:         0x01,
				BNumDescriptors:      0x03,
				BClassDescriptorType: 0x22,
				WDescriptorLength:    0x003F,
				OptionalDescriptorTypes: []hidprotocol.OptionalHIDDescriptorTypes{
					{
						BOptionalDescriptorType:   descriptor.DESCRIPTOR_TYPE_HID_REPORT,
						BOptionalDescriptorLength: 0x0041,
					},
					{
						BOptionalDescriptorType:   descriptor.DESCRIPTOR_TYPE_HID_REPORT,
						BOptionalDescriptorLength: 0x0042,
					},
				},
			},
			err: nil,
		},
		{
			name: "Error_ControlTransferError",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2100), uint16(1), make([]byte, hidprotocol.HID_DESCRIPTOR_LENGTH)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							return 0, errControl
						})

					return mocks
				},
				config: config,
			},
			desc: hidprotocol.HIDDescriptor{},
			err:  errControl,
		},
		{
			name: "Error_ControlTransferError2",
			fields: fields{
				mocks: func(ctrl *gomock.Controller) mocks {
					mocks := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)
					readData := []byte{
						0x0c,
						0x21,
						0x10, 0x01,
						0x01,
						0x03,
						0x22,
						0x3F, 0x00,
						0x22,
						0x41, 0x00,
						0x22,
						0x42, 0x00,
					}

					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2100), uint16(1), make([]byte, hidprotocol.HID_DESCRIPTOR_LENGTH)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							copy(data, readData)

							return len(data), nil
						})
					mocks.device.EXPECT().
						Control(uint8(0b1000_0001), uint8(0x06), uint16(0x2100), uint16(1), make([]byte, hidprotocol.HID_DESCRIPTOR_LENGTH+6)).
						DoAndReturn(func(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
							return 0, errControl
						})

					return mocks
				},
				config: config,
			},
			desc: hidprotocol.HIDDescriptor{},
			err:  errControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := slog.Default()
			mockUSBs := test.fields.mocks(ctrl)

			hidDevice, err := hid.NewDevice(mockUSBs.device, test.fields.config, logger)
			assert.NoError(t, err)

			err = hidDevice.SetTarget(1, 1, 0)
			assert.NoError(t, err)

			desc, err := hidDevice.GetHIDDescriptor()
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.desc, desc)
		})
	}
}

func TestDevice_GetStringDescriptor(t *testing.T) {
	stringDesc := "Made by ntch.dev"
	ctrl := gomock.NewController(t)
	logger := slog.Default()
	mockUSBs := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

	mockUSBs.device.EXPECT().GetStringDescriptor(4).Return(stringDesc, nil)

	hidDevice, err := hid.NewDevice(mockUSBs.device, config, logger)
	assert.NoError(t, err)

	err = hidDevice.SetTarget(1, 1, 0)
	assert.NoError(t, err)

	str, err := hidDevice.GetStringDescriptor(4)
	assert.NoError(t, err)
	assert.Equal(t, stringDesc, str)
}

func TestDevice_GetDeviceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := slog.Default()
	mockUSBs := createSetupTargetMocks(ctrl, 1, 1, 0, 1, 1)

	hidDevice, err := hid.NewDevice(mockUSBs.device, config, logger)
	assert.NoError(t, err)

	err = hidDevice.SetTarget(1, 1, 0)
	assert.NoError(t, err)

	info := hidDevice.GetDeviceInfo()

	assert.Equal(t, deviceDesc, info.DeviceDesc)
	assert.Equal(t, 1, info.GetConfigNumber())
	assert.Equal(t, 1, info.GetInterfaceNumber())
	assert.Equal(t, 0, info.GetAltSettingNumber())
}
