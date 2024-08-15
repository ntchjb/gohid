package hid

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/gousb"
	"github.com/ntchjb/usbip-virtual-device/usb/protocol/hid"
	"github.com/ntchjb/usbip-virtual-device/usb/protocol/hid/report"
)

const (
	HID_CLASS_ID        uint8  = 3
	HID_MAX_REPORT_SIZE uint16 = 4096
)

type SubClass uint8

const (
	HID_SUBCLASS_BOOT_INTERFACE SubClass = 1
	HID_SUBCLASS_NONE           SubClass = 0
)

type HIDProtocol uint8

const (
	HID_PROTOCOL_NONE     HIDProtocol = 0
	HID_PROTOCOL_KEYBOARD HIDProtocol = 1
	HID_PROTOCOL_MOUSE    HIDProtocol = 2
)

type SetupRequestType uint8

const (
	SETUP_REQUEST_TYPE_STANDARD SetupRequestType = 0x00 << 5
	SETUP_REQUEST_TYPE_CLASS    SetupRequestType = 0x01 << 5
	SETUP_REQUEST_TYPE_VENDOR   SetupRequestType = 0x02 << 5
	SETUP_REQUEST_TYPE_RESERVED SetupRequestType = 0x03 << 5
)

type SetupRequestRecipient uint8

const (
	SETUP_RECIPIENT_DEVICE    SetupRequestRecipient = 0x00
	SETUP_RECIPIENT_INTERFACE SetupRequestRecipient = 0x01
	SETUP_RECIPIENT_ENDPOINT  SetupRequestRecipient = 0x02
	SETUP_RECIPIENT_OTHER     SetupRequestRecipient = 0x03
)

type SetupEndpointDirection uint8

const (
	SETUP_EP_DIR_OUT SetupEndpointDirection = 0x00
	SETUP_EP_DIR_IN  SetupEndpointDirection = 0x80
)

type SetupRequest uint8

const (
	SETUP_REQUEST_GET_DESCRIPTOR SetupRequest = 0x06

	SETUP_REQUEST_HID_GET_REPORT   SetupRequest = 0x01
	SETUP_REQUEST_HID_GET_IDLE     SetupRequest = 0x02
	SETUP_REQUEST_HID_GET_PROTOCOL SetupRequest = 0x03
	SETUP_REQUEST_HID_SET_REPORT   SetupRequest = 0x09
	SETUP_REQUEST_HID_SET_IDLE     SetupRequest = 0x0A
	SETUP_REQUEST_HID_SET_PROTOCOL SetupRequest = 0x0B
)

type ReportType uint8

const (
	REPORT_TYPE_INPUT   ReportType = 0x01
	REPORT_TYPE_OUTPUT  ReportType = 0x02
	REPORT_TYPE_FEATURE ReportType = 0x03
)

type EndpointTransferType uint8

// Only two of these transfer types are used for HID class devices
const (
	ENDPOINT_TRANSFER_TYPE_CONTROL   EndpointTransferType = 0
	ENDPOINT_TRANSFER_TYPE_INTERRUPT EndpointTransferType = 3
)

type ClassDescriptorType uint8

const (
	DESCRIPTOR_TYPE_HID      ClassDescriptorType = 0x21
	DESCRIPTOR_TYPE_REPORT   ClassDescriptorType = 0x22
	DESCRIPTOR_TYPE_PHYSICAL ClassDescriptorType = 0x23
)

type DeviceEndpoint struct {
	// Endpoint number to be used for connection
	Number int
	// Endpoint address described by endpoint descriptor
	// The address is encoded as follows:
	//
	// Bit 0..3 -> The endpoint number,
	// Bit 4..6 -> Reserved, reset to zero,
	// Bit 7 -> Direction (ignored for Control endpoints):
	// 0 - OUT endpoint,
	// 1 - IN endpoint
	Address uint8
	// Endpoint direction: true -> IN, false -> OUT
	Direction bool
	// Maximum packet size this endpoint is capable of
	// sending or receiving when this configuration is
	// selected.
	// For interrupt endpoints, this value is used to reserve the
	// bus time in the schedule, required for the per frame data
	// payloads. Smaller data payloads may be sent, but will
	// terminate the transfer and thus require intervention to
	// restart.
	MaxPacketSize int
	// Interval for polling endpoint for data transfers. Expressed in milliseconds.
	PollInterval time.Duration
	// Transfer types of given endpoint
	TransferType EndpointTransferType
}

type DeviceInfo struct {
	DeviceDesc gousb.DeviceDesc
	ConfigDesc gousb.ConfigDesc
	// Connection IDs used by this library
	// 0: Configuration number
	// 1: Interface number
	// 2: Setting number
	// 3: Setting Alternate number
	Target [3]int
	// HID subclass
	Subclass SubClass
	// HID protocol
	Protocol HIDProtocol
	// List of endpoints
	Endpoints []DeviceEndpoint
}

func (d DeviceInfo) String() string {
	var builder strings.Builder
	builder.WriteRune('[')
	builder.WriteString(d.DeviceDesc.Vendor.String())
	builder.WriteRune(':')
	builder.WriteString(d.DeviceDesc.Product.String())
	builder.WriteString("] Conf #")
	builder.WriteString(strconv.Itoa(d.Target[0]))
	builder.WriteString(" Intf #")
	builder.WriteString(strconv.Itoa(d.Target[1]))
	builder.WriteString(" Sett #")
	builder.WriteString(strconv.Itoa(d.Target[2]))
	builder.WriteString(" Speed: ")
	builder.WriteString(d.DeviceDesc.Speed.String())
	builder.WriteString(", CtrlSize: ")
	builder.WriteString(strconv.Itoa(d.DeviceDesc.MaxControlPacketSize))

	builder.WriteString(", Ep:[")
	for i, endpoint := range d.Endpoints {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("#" + strconv.Itoa(endpoint.Number))
		if endpoint.Direction {
			builder.WriteString("(IN)")
		} else {
			builder.WriteString("(OUT)")
		}
	}
	builder.WriteRune(']')

	return builder.String()
}

type DeviceInfos []DeviceInfo

func (d DeviceInfos) String() string {
	var res string
	for _, info := range d {
		res += info.String() + "\n"
	}
	return res
}

type DeviceAccessor interface {
	// Close device connection
	Close() error
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
	GetReportDescriptor() (report.HIDReportDescriptor, error)
	// Get device info
	GetDeviceInfo() DeviceInfo
	// Get HID descriptor using Get_Descriptor transfer, via control endpoint
	GetHIDDescriptor() (hid.HIDDescriptor, error)
}

type DeviceOpener interface {
	Open() (DeviceAccessor, error)
}
