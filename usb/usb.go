package usb

import (
	"context"

	"github.com/google/gousb"
)

// StreamReader provides ability to write data to USB device via endpoint with higher throughput
type StreamWriter interface {
	// Write data to USB device via endpoint OUT
	WriteContext(ctx context.Context, data []byte) (int, error)
	// Close stream writer system
	Close() error
}

// StreamReader provides ability to read data from USB device via endpoint with higher throughput
type StreamReader interface {
	// Read data from USB device via endpoint IN
	ReadContext(ctx context.Context, data []byte) (int, error)
	// Close stream reader system
	Close() error
}

// OutEndpoint represents endpoint OUT of USB device
type OutEndpoint interface {
	// Get endpoint descriptor
	Descriptor() gousb.EndpointDesc
	// Create stream writer to write data to OUT endpoint
	NewStream(count int) (StreamWriter, error)
}

// OutEndpoint represents endpoint IN of USB device
type InEndpoint interface {
	// Get endpoint descriptor
	Descriptor() gousb.EndpointDesc
	// Create stream reader to read data from IN endpoint
	NewStream(count int) (StreamReader, error)
}

// Interface represents interface of USB device located under USB device's configuration
type Interface interface {
	// Un-claim USB device's interface and release associated resources
	Close() error
	// Get IN endpoint by endpoint number
	InEndpoint(num int) (InEndpoint, error)
	// Get OUT endpoint by endpoint number
	OutEndpoint(num int) (OutEndpoint, error)
}

type Config interface {
	// Get interface by given number and alt
	Interface(num, alt int) (Interface, error)
	// Un-claim USB device's configuration and release associated resources
	Close() error
}

// Device represents an opened connection of a USB device.
type Device interface {
	// Allow this library to automatically detach this device from kernel when need to claim device's resource,
	// such as device interfaces, and attach back when the resource is released
	SetAutoDetach(autoDetach bool) error
	// Claim a configuration of USB device
	Config(configNumber int) (Config, error)
	// Get device descriptor
	Descriptor() *gousb.DeviceDesc
	// Send a USB device request via control endpoint
	Control(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error)
	// Get string descriptor of USB device by sending GET_DESCRIPTOR request.
	// The requerst is sent via control endpoint.
	GetStringDescriptor(index int) (string, error)
	// Get manufacturer string by getting string descriptor from USB device
	Manufacturer() (string, error)
	// Get product name string by getting string descriptor from USB device
	Product() (string, error)
	// Get device serial number string by getting string descriptor from USB device
	SerialNumber() (string, error)
	// Close USB device connection and release resources
	Close() error
}

// Context represents connections between this library and USB devices. It is capable of providing a list of connected USB devices,
// open USB devices, etc.
type Context interface {
	// Iterate through list of connected USB devices and get their descriptors from the devices
	IterateDevices(reader func(desc *gousb.DeviceDesc)) error
	// Open a USB device by given vendor ID and product ID
	OpenDevice(vid, pid gousb.ID) (Device, error)
	// Close context and release all associated resources
	Close() error
}
