package hid

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/gousb"
	"github.com/ntchjb/gohid/usb"
	"github.com/ntchjb/usbip-virtual-device/usb/protocol/hid"
	hidreport "github.com/ntchjb/usbip-virtual-device/usb/protocol/hid/report"
)

var (
	ErrEmptyData             = errors.New("empty data")
	ErrUninitializedDevice   = errors.New("uninitialized device")
	ErrUninitializedEndpoint = errors.New("uninitialized endpoint")
	ErrDeviceIsNil           = errors.New("device is nil")
	ErrEndpointInNotFound    = errors.New("endpoint IN not found")
)

const (
	DEFAULT_ENDPOINT_STREAM_COUNT = 16
)

type DeviceConfig struct {
	// Number of stream lanes used for streaming data on interrupt endpoints
	StreamLaneCount int
}

type Device interface {
	// Set profile to the device on which configuration/interface/alternateSetting to be used
	SetTarget(confNumber, infNumber, altNumber int) error
	// Close device connection
	Close() error
	// Set an HID device to automatically detached from kernel when claim its interfaces
	SetAutoDetach(autoDetach bool) error
	// Write an Output report to HID device, via interrupt OUT endpoint
	WriteOutput(ctx context.Context, data []byte) (int, error)
	// Read an Input report from a HID device, via interrupt IN endpoint
	ReadInput(ctx context.Context, data []byte) (int, error)
	// Send a Feature Report using Set_Report transfer, via control endpoint
	// The first byte of data must contain the Report ID. For device that support single report type, set it to 0x00
	SendFeatureReport(data []byte) (int, error)
	// Get a Feature report from a HID device using Get_Report transfer, via control endpoint
	GetFeatureReport(data []byte) (int, error)
	// Send Output Report to HID device using Set_Report transfer, via control endpoint
	SendOutputReport(data []byte) (int, error)
	// Get Input report from HID device using Get_Report transfer, via control endpoint
	GetInputReport(data []byte) (int, error)
	// Get device serial number using Get_Descriptor transfer (indexed string), via control endpoint
	GetSerialNumber() (string, error)
	// Get device product name using Get_Descriptor transfer (indexed string), via control endpoint
	GetProduct() (string, error)
	// Get device manufacturer using Get_Descriptor transfer (indexed string), via control endpoint
	GetManufacturer() (string, error)
	// Get report descriptor using Get_Descriptor transfer, via control endpoint
	GetReportDescriptor() (hidreport.HIDReportDescriptor, error)
	// Get HID descriptor using Get_Descriptor transfer, via control endpoint
	GetHIDDescriptor() (hid.HIDDescriptor, error)
	// Get string descriptor
	GetStringDescriptor(index int) (string, error)
	// Get device info
	GetDeviceInfo() DeviceInfo
}

func NewDevice(device usb.Device, config DeviceConfig, logger *slog.Logger) (Device, error) {
	if device == nil {
		return nil, ErrDeviceIsNil
	}

	return &deviceImpl{
		device:  device,
		dConfig: config,
		logger:  logger,
	}, nil
}

type deviceImpl struct {
	device usb.Device
	config usb.Config
	intf   usb.Interface
	writer usb.StreamWriter
	reader usb.StreamReader

	dConfig DeviceConfig

	deviceInfo DeviceInfo
	logger     *slog.Logger
}

func (d *deviceImpl) SetAutoDetach(autoDetach bool) error {
	// This is important to allow this library to attach device's interfaces.
	// Without this call, manual detach of the device from kernel is required
	// to successfully claim device interfaces.
	if err := d.device.SetAutoDetach(autoDetach); err != nil {
		desc := d.device.Descriptor()
		return fmt.Errorf("unable to set auto detach for device %v:%v: %w", desc.Vendor, desc.Product, err)
	}

	return nil
}

func (d *deviceImpl) SetTarget(confNumber, infNumber, altNumber int) error {
	var err error
	var deviceInfo DeviceInfo
	var cfg usb.Config
	var intf usb.Interface
	var epIn usb.InEndpoint
	var epOut usb.OutEndpoint
	var writer usb.StreamWriter
	var reader usb.StreamReader

	defer func() {
		// All opened connections should be closed if error occurred
		if err != nil {
			if intf != nil {
				intf.Close()
			}
			if cfg != nil {
				cfg.Close()
			}
		}
	}()

	deviceDesc := d.device.Descriptor()
	if err = deviceInfo.FromDeviceDesc(deviceDesc, confNumber, infNumber, altNumber); err != nil {
		return fmt.Errorf("unable to gain device info from device: %w", err)
	}

	if d.writer != nil {
		if err := d.writer.Close(); err != nil {
			d.logger.Error("unable to close existing stream writer", "err", err)
		}
	}
	if d.reader != nil {
		if err := d.reader.Close(); err != nil {
			d.logger.Error("unable to close existing stream reader", "err", err)
		}
	}
	if d.intf != nil {
		if err := d.intf.Close(); err != nil {
			d.logger.Error("unable to close existing interface", "err", err)
		}
	}
	if d.config != nil {
		if err := d.config.Close(); err != nil {
			d.logger.Error("unable to close existing config", "err", err)
		}
	}

	cfg, err = d.device.Config(confNumber)
	if err != nil {
		return fmt.Errorf("unable to get config #%d for device %v:%v: %w", confNumber, deviceDesc.Vendor, deviceDesc.Product, err)
	}
	logger := d.logger.With("cfg", confNumber)
	intf, err = cfg.Interface(infNumber, altNumber)
	if err != nil {
		return fmt.Errorf("unable to get interface #%d:%d for config #%d of device %v:%v: %w", infNumber, altNumber, confNumber, deviceDesc.Vendor, deviceDesc.Product, err)
	}
	logger = logger.With("intf", infNumber, "alt", altNumber)
	endpoints := deviceInfo.GetEndpoints()
	for _, endpoint := range endpoints {
		if endpoint.Direction == gousb.EndpointDirectionIn && epIn == nil {
			logger.Info("use endpoint IN", "number", endpoint.Number)
			epIn, err = intf.InEndpoint(endpoint.Number)
			if err != nil {
				return fmt.Errorf("unable to get IN endpoint at #%d for device %04x:%04x: %w", endpoint.Number, deviceDesc.Vendor, deviceDesc.Product, err)
			}
		} else if endpoint.Direction == gousb.EndpointDirectionOut && epOut == nil {
			logger.Info("use endpoint OUT", "number", endpoint.Number)
			epOut, err = intf.OutEndpoint(endpoint.Number)
			if err != nil {
				return fmt.Errorf("unable to get OUT endpoint at #%d for device %04x:%04x: %w", endpoint.Number, deviceDesc.Vendor, deviceDesc.Product, err)
			}
		}
	}

	if epIn == nil {
		err = fmt.Errorf("endpoint IN not found for the device %04x:%04x, conf #%d, inf #%d, alt #%d: %w", deviceDesc.Vendor, deviceDesc.Product, confNumber, infNumber, altNumber, ErrEndpointInNotFound)
		return err
	}
	reader, err = epIn.NewStream(d.dConfig.StreamLaneCount)
	if err != nil {
		return fmt.Errorf("unable to create stream reader for endpoint %d: %w", epIn.Descriptor().Number, err)
	}
	if epOut != nil {
		writer, err = epOut.NewStream(d.dConfig.StreamLaneCount)
		if err != nil {
			return fmt.Errorf("unable to create stream writer for endpoint %d: %w", epOut.Descriptor().Number, err)
		}
	}

	d.config = cfg
	d.intf = intf
	d.writer = writer
	d.reader = reader
	d.deviceInfo = deviceInfo

	return nil
}

func (d *deviceImpl) Close() error {
	var allErrs error
	if d.reader != nil {
		if err := d.reader.Close(); err != nil {
			allErrs = errors.Join(allErrs, fmt.Errorf("unable to close stream reader: %w", err))
		}
	}
	if d.writer != nil {
		if err := d.writer.Close(); err != nil {
			allErrs = errors.Join(allErrs, fmt.Errorf("unable to close stream writer: %w", err))
		}
	}
	if d.intf != nil {
		if err := d.intf.Close(); err != nil {
			allErrs = errors.Join(allErrs, fmt.Errorf("unable to close interface: %w", err))
		}
	}
	if d.config != nil {
		if err := d.config.Close(); err != nil {
			allErrs = errors.Join(allErrs, fmt.Errorf("unable to close selected device configuration #%d: %w", d.deviceInfo.GetConfigNumber(), err))
		}
	}
	if d.device != nil {
		if err := d.device.Close(); err != nil {
			allErrs = errors.Join(allErrs, fmt.Errorf("unable to close device at %04x:%04x: %w", d.deviceInfo.DeviceDesc.Vendor, d.deviceInfo.DeviceDesc.Product, err))
		}
	}

	return allErrs
}

func (d *deviceImpl) WriteOutput(ctx context.Context, data []byte) (int, error) {
	var isSkippedReportID bool
	if d.writer == nil {
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

	byteWritten, err := d.writer.WriteContext(ctx, data)
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

	byteRead, err := d.reader.ReadContext(ctx, data)
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

	reportNumber := data[0]
	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}
	byteSend, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_CLASS)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_OUT),
		uint8(SETUP_REQUEST_HID_SET_REPORT),
		(uint16(REPORT_TYPE_FEATURE)<<8)|uint16(reportNumber),
		uint16(d.deviceInfo.GetInterfaceNumber()),
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
		return 0, ErrEmptyData
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
		uint16(d.deviceInfo.GetInterfaceNumber()),
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
	reportNumber := data[0]
	if reportNumber == 0x00 {
		data = data[1:]
		isSkippedReportID = true
	}

	byteSend, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_CLASS)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_OUT),
		uint8(SETUP_REQUEST_HID_SET_REPORT),
		(uint16(REPORT_TYPE_OUTPUT)<<8)|uint16(reportNumber),
		uint16(d.deviceInfo.GetInterfaceNumber()),
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
		return 0, ErrEmptyData
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
		uint16(d.deviceInfo.GetInterfaceNumber()),
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
	buf := make([]byte, HID_MAX_REPORT_SIZE)

	n, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_STANDARD)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_GET_DESCRIPTOR),
		(uint16(DESCRIPTOR_TYPE_REPORT)<<8)|uint16(0), // Descriptor Index is zero for all HID descriptors except Physical descriptors
		uint16(d.deviceInfo.GetInterfaceNumber()),
		buf,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get report descriptor via control endpoint: %w", err)
	}

	return buf[:n], nil
}

func (d *deviceImpl) GetHIDDescriptor() (hid.HIDDescriptor, error) {
	var desc hid.HIDDescriptor

	// #1: Get partial data first to know the whole data size
	data := make([]byte, hid.HID_DESCRIPTOR_LENGTH)

	_, err := d.device.Control(
		uint8(SETUP_REQUEST_TYPE_STANDARD)|uint8(SETUP_RECIPIENT_INTERFACE)|uint8(SETUP_EP_DIR_IN),
		uint8(SETUP_REQUEST_GET_DESCRIPTOR),
		(uint16(DESCRIPTOR_TYPE_HID)<<8)|uint16(0), // Descriptor Index is zero
		uint16(d.deviceInfo.GetInterfaceNumber()),
		data,
	)

	if err != nil {
		return desc, fmt.Errorf("unable send input report via control endpoint: %w", err)
	}
	if err := desc.Decode(bytes.NewBuffer(data)); err != nil && !errors.Is(err, io.EOF) {
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
		uint16(d.deviceInfo.GetInterfaceNumber()),
		data,
	)

	if err != nil {
		return hid.HIDDescriptor{}, fmt.Errorf("unable send input report via control endpoint (2nd time): %w", err)
	}
	if err := desc.Decode(bytes.NewBuffer(data)); err != nil {
		return hid.HIDDescriptor{}, fmt.Errorf("unable to decode HID descriptor (2nd time): %w", err)
	}

	return desc, nil
}

func (d *deviceImpl) GetStringDescriptor(index int) (string, error) {
	return d.device.GetStringDescriptor(index)
}
