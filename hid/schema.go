package hid

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
