package usb

import (
	"context"

	"github.com/google/gousb"
)

type StreamWriter interface {
	WriteContext(ctx context.Context, data []byte) (int, error)
	Close() error
}

type StreamReader interface {
	ReadContext(ctx context.Context, data []byte) (int, error)
	Close() error
}

type OutEndpoint interface {
	Descriptor() gousb.EndpointDesc
	NewStream(count int) (StreamWriter, error)
}

type InEndpoint interface {
	Descriptor() gousb.EndpointDesc
	NewStream(count int) (StreamReader, error)
}

type Interface interface {
	Close() error
	InEndpoint(num int) (InEndpoint, error)
	OutEndpoint(num int) (OutEndpoint, error)
}

type Config interface {
	Interface(num, alt int) (Interface, error)
	Close() error
}

type Device interface {
	SetAutoDetach(autoDetach bool) error
	Config(configNumber int) (Config, error)
	Descriptor() *gousb.DeviceDesc
	Control(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error)
	GetStringDescriptor(index int) (string, error)
	Manufacturer() (string, error)
	Product() (string, error)
	SerialNumber() (string, error)

	Close() error
}

type Context interface {
	IterateDevices(reader func(desc *gousb.DeviceDesc)) error
	OpenDevice(vid, pid gousb.ID) (Device, error)
	Close() error
}
