package hid

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/gousb"
	"github.com/ntchjb/usbip-virtual-device/usb/protocol/hid"
	hidreport "github.com/ntchjb/usbip-virtual-device/usb/protocol/hid/report"
)

var (
	ErrEmptyData             = errors.New("empty data")
	ErrUninitializedDevice   = errors.New("uninitialized device")
	ErrUninitializedEndpoint = errors.New("uninitialized endpoint")
)

func NewDevice(device *gousb.Device, deviceInfo DeviceInfo) (DeviceAccessor, error) {
	var err error
	var cfg *gousb.Config
	var intf *gousb.Interface
	var epIn *gousb.InEndpoint
	var epOut *gousb.OutEndpoint

	defer func() {
		// All opened connections should be closed if error occurred
		if err != nil {
			if intf != nil {
				intf.Close()
			}
			if cfg != nil {
				cfg.Close()
			}
			if device != nil {
				device.Close()
			}
		}
	}()

	if err := device.SetAutoDetach(true); err != nil {
		return nil, fmt.Errorf("unable to set auto detach for device %v:%v: %w", deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product, err)
	}

	cfg, err = device.Config(deviceInfo.Target[0])
	if err != nil {
		return nil, fmt.Errorf("unable to get config #%d for device %v:%v: %w", deviceInfo.Target[0], deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product, err)
	}
	intf, err = cfg.Interface(deviceInfo.Target[1], deviceInfo.Target[2])
	if err != nil {
		return nil, fmt.Errorf("unable to get interface #%d:%d for config #%d of device %v:%v: %w", deviceInfo.Target[1], deviceInfo.Target[2], deviceInfo.Target[0], deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product, err)
	}

	for _, endpoint := range deviceInfo.Endpoints {
		if endpoint.Direction == bool(gousb.EndpointDirectionIn) && epIn == nil {
			epIn, err = intf.InEndpoint(endpoint.Number)
			if err != nil {
				return nil, fmt.Errorf("unable to get IN endpoint at #%d for device %04x:%04x: %w", endpoint.Number, deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product, err)
			}
		} else if endpoint.Direction == bool(gousb.EndpointDirectionOut) && epOut == nil {
			epOut, err = intf.OutEndpoint(endpoint.Number)
			if err != nil {
				return nil, fmt.Errorf("unable to get OUT endpoint at #%d for device %04x:%04x: %w", endpoint.Number, deviceInfo.DeviceDesc.Vendor, deviceInfo.DeviceDesc.Product, err)
			}
		}
	}

	return &deviceImpl{
		device: device,
		config: cfg,
		intf:   intf,
		epIn:   epIn,
		epOut:  epOut,
	}, nil
}

type deviceImpl struct {
	device *gousb.Device
	config *gousb.Config
	intf   *gousb.Interface
	epIn   *gousb.InEndpoint
	epOut  *gousb.OutEndpoint

	deviceInfo DeviceInfo
}

func (d *deviceImpl) Close() error {
	if d.intf != nil {
		d.intf.Close()
	}

	if d.config != nil {
		if err := d.config.Close(); err != nil {
			return fmt.Errorf("unable to close selected device configuration #%d at %04x:%04x: %w", d.deviceInfo.Target[0], d.device.Desc.Vendor, d.device.Desc.Product, err)
		}
	}
	if d.device != nil {
		if err := d.device.Close(); err != nil {
			return fmt.Errorf("unable to close device at %04x:%04x: %w", d.device.Desc.Vendor, d.device.Desc.Product, err)
		}
	}

	return nil
}

func (d *deviceImpl) WriteOutput(ctx context.Context, data []byte) (int, error) {
	var isSkippedReportID bool
	if d.epOut == nil {
		return d.SendOutputReport(data)
	}
	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	reportNumber := data[0]

	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}

	byteWritten, err := d.epOut.WriteContext(ctx, data)
	if err != nil {
		return byteWritten, fmt.Errorf("unable to write output report to interrupt OUT endpoint: %w", err)
	}

	if isSkippedReportID {
		byteWritten += 1
	}

	return byteWritten, nil
}

func (d *deviceImpl) ReadInput(ctx context.Context, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if d.epIn == nil {
		return 0, ErrUninitializedEndpoint
	}
	byteRead, err := d.epIn.ReadContext(ctx, data)
	if err != nil {
		return byteRead, fmt.Errorf("unable to read report from interrupt IN endpoint: %w", err)
	}

	return byteRead, nil
}

func (d *deviceImpl) SendFeatureReport(data []byte) (int, error) {
	var isSkippedReportID bool

	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	if d.device == nil {
		return 0, ErrUninitializedDevice
	}

	reportNumber := data[0]
	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}
	byteSend, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_CLASS)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_OUT),
		uint8(SETUP_REQUEST_HID_SET_REPORT),
		(uint16(REPORT_TYPE_FEATURE)<<8)|uint16(reportNumber),
		uint16(d.deviceInfo.Target[1]),
		data,
	)

	if err != nil {
		return 0, fmt.Errorf("unable set feature report via control endpoint: %w", err)
	}

	if isSkippedReportID {
		byteSend++
	}

	return byteSend, nil
}

func (d *deviceImpl) GetFeatureReport(data []byte) (int, error) {
	var isSkippedReportID bool

	if len(data) == 0 {
		return 0, nil
	}
	if d.device == nil {
		return 0, ErrUninitializedDevice
	}
	reportNumber := data[0]
	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}

	byteSend, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_CLASS)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_HID_GET_REPORT),
		(uint16(REPORT_TYPE_FEATURE)<<8)|uint16(reportNumber),
		uint16(d.deviceInfo.Target[1]),
		data,
	)

	if err != nil {
		return 0, fmt.Errorf("unable get feature report via control endpoint: %w", err)
	}

	if isSkippedReportID {
		byteSend++
	}

	return byteSend, nil
}

func (d *deviceImpl) SendOutputReport(data []byte) (int, error) {
	var isSkippedReportID bool

	if len(data) == 0 {
		return 0, ErrEmptyData
	}
	if d.device == nil {
		return 0, ErrUninitializedDevice
	}
	reportNumber := data[0]
	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}

	byteSend, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_CLASS)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_OUT),
		uint8(SETUP_REQUEST_HID_SET_REPORT),
		(uint16(REPORT_TYPE_OUTPUT)<<8)|uint16(reportNumber),
		uint16(d.deviceInfo.Target[1]),
		data,
	)

	if err != nil {
		return 0, fmt.Errorf("unable send output report via control endpoint: %w", err)
	}

	if isSkippedReportID {
		byteSend++
	}

	return byteSend, nil
}

func (d *deviceImpl) GetInputReport(data []byte) (int, error) {
	var isSkippedReportID bool

	if len(data) == 0 {
		return 0, nil
	}
	if d.device == nil {
		return 0, ErrUninitializedDevice
	}
	reportNumber := data[0]
	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}

	byteSend, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_CLASS)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_HID_GET_REPORT),
		(uint16(REPORT_TYPE_INPUT)<<8)|uint16(reportNumber),
		uint16(d.deviceInfo.Target[1]),
		data,
	)

	if err != nil {
		return 0, fmt.Errorf("unable send input report via control endpoint: %w", err)
	}

	if isSkippedReportID {
		byteSend++
	}

	return byteSend, nil
}

func (d *deviceImpl) GetManufacturer() (string, error) {
	return d.device.Manufacturer()
}

func (d *deviceImpl) GetProduct() (string, error) {
	return d.device.Product()
}

func (d *deviceImpl) GetSerialNumber() (string, error) {
	return d.device.SerialNumber()
}

func (d *deviceImpl) GetDeviceInfo() DeviceInfo {
	return d.deviceInfo
}

func (d *deviceImpl) GetReportDescriptor() (hidreport.HIDReportDescriptor, error) {
	if d.device == nil {
		return nil, ErrUninitializedDevice
	}

	buf := make([]byte, HID_MAX_REPORT_SIZE)

	_, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_STANDARD)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_GET_DESCRIPTOR),
		(uint16(DESCRIPTOR_TYPE_REPORT)<<8)|uint16(0), // Descriptor Index is zero
		uint16(d.deviceInfo.Target[1]),
		buf,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get report descriptor via control endpoint: %w", err)
	}

	return buf, nil
}

func (d *deviceImpl) GetHIDDescriptor() (hid.HIDDescriptor, error) {
	var desc hid.HIDDescriptor
	if d.device == nil {
		return desc, ErrUninitializedDevice
	}

	// #1: Get partial data first to know the whole data size
	data := make([]byte, hid.HID_DESCRIPTOR_LENGTH)

	_, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_STANDARD)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_GET_DESCRIPTOR),
		(uint16(DESCRIPTOR_TYPE_HID)<<8)|uint16(0), // Descriptor Index is zero
		uint16(d.deviceInfo.Target[1]),
		data,
	)

	if err != nil {
		return desc, fmt.Errorf("unable send input report via control endpoint: %w", err)
	}
	if err := desc.Decode(bytes.NewBuffer(data)); err != nil {
		return desc, fmt.Errorf("unable to decode HID descriptor: %w", err)
	}

	if desc.BNumDescriptors == 1 {
		return desc, nil
	}

	// #2: Now get the whole descriptor data, if any
	data = make([]byte, hid.HID_DESCRIPTOR_LENGTH+(desc.BNumDescriptors-1)*3)
	_, err = d.device.Control(
		uint8(SETUP_REQUEST_TYPE_STANDARD)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_GET_DESCRIPTOR),
		(uint16(DESCRIPTOR_TYPE_HID)<<8)|uint16(0), // Descriptor Index is zero
		uint16(d.deviceInfo.Target[1]),
		data,
	)

	if err != nil {
		return desc, fmt.Errorf("unable send input report via control endpoint (2nd time): %w", err)
	}
	if err := desc.Decode(bytes.NewBuffer(data)); err != nil {
		return desc, fmt.Errorf("unable to decode HID descriptor (2nd time): %w", err)
	}

	return desc, nil
}
