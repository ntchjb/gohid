package usb

import (
	"context"
	"errors"

	"github.com/google/gousb"
)

var (
	ErrGOUSBDeviceIsNil = errors.New("gousb device is nil")
)

type gousbStreamWriter struct {
	stream *gousb.WriteStream
}

func NewGOUSBStreamWriter(stream *gousb.WriteStream) StreamWriter {
	return &gousbStreamWriter{
		stream: stream,
	}
}

func (g *gousbStreamWriter) WriteContext(ctx context.Context, data []byte) (int, error) {
	return g.stream.WriteContext(ctx, data)
}

func (g *gousbStreamWriter) Close() error {
	return g.stream.Close()
}

type gousbStreamReader struct {
	stream *gousb.ReadStream
}

func NewGOUSBStreamReader(stream *gousb.ReadStream) StreamReader {
	return &gousbStreamReader{
		stream: stream,
	}
}

func (g *gousbStreamReader) ReadContext(ctx context.Context, data []byte) (int, error) {
	return g.stream.ReadContext(ctx, data)
}

func (g *gousbStreamReader) Close() error {
	return g.stream.Close()
}

type gousbOutEndpoint struct {
	ep *gousb.OutEndpoint
}

func NewGOUSBOutEndpoint(ep *gousb.OutEndpoint) OutEndpoint {
	return &gousbOutEndpoint{
		ep: ep,
	}
}

func (g *gousbOutEndpoint) NewStream(count int) (StreamWriter, error) {
	writeStream, err := g.ep.NewStream(g.ep.Desc.MaxPacketSize, count)
	if err != nil {
		return nil, err
	}

	return NewGOUSBStreamWriter(writeStream), nil
}

func (g *gousbOutEndpoint) Descriptor() gousb.EndpointDesc {
	return g.ep.Desc
}

type gousbInEndpoint struct {
	ep *gousb.InEndpoint
}

func NewGOUSBInEndpoint(ep *gousb.InEndpoint) InEndpoint {
	return &gousbInEndpoint{
		ep: ep,
	}
}

func (g *gousbInEndpoint) NewStream(count int) (StreamReader, error) {
	readStream, err := g.ep.NewStream(g.ep.Desc.MaxPacketSize, count)
	if err != nil {
		return nil, err
	}

	return NewGOUSBStreamReader(readStream), nil
}

func (g *gousbInEndpoint) Descriptor() gousb.EndpointDesc {
	return g.ep.Desc
}

type gousbInterface struct {
	intf *gousb.Interface
}

func NewGOUSBInterface(intf *gousb.Interface) Interface {
	return &gousbInterface{
		intf: intf,
	}
}

func (g *gousbInterface) Close() error {
	g.intf.Close()
	return nil
}

func (g *gousbInterface) InEndpoint(num int) (InEndpoint, error) {
	inEndpoint, err := g.intf.InEndpoint(num)
	if err != nil {
		return nil, err
	}

	return NewGOUSBInEndpoint(inEndpoint), nil
}
func (g *gousbInterface) OutEndpoint(num int) (OutEndpoint, error) {
	outEndpoint, err := g.intf.OutEndpoint(num)
	if err != nil {
		return nil, err
	}

	return NewGOUSBOutEndpoint(outEndpoint), nil
}

type gousbConfig struct {
	config *gousb.Config
}

func NewGOUSBConfiguration(config *gousb.Config) Config {
	return &gousbConfig{
		config: config,
	}
}

func (g *gousbConfig) Interface(num, alt int) (Interface, error) {
	intf, err := g.config.Interface(num, alt)
	if err != nil {
		return nil, err
	}

	return NewGOUSBInterface(intf), nil
}

func (g *gousbConfig) Close() error {
	return g.config.Close()
}

type gousbDevice struct {
	device *gousb.Device
}

func NewGOUSBDevice(device *gousb.Device) (Device, error) {
	if device == nil {
		return nil, ErrGOUSBDeviceIsNil
	}
	return &gousbDevice{device: device}, nil
}

func (g *gousbDevice) SetAutoDetach(autodetach bool) error {
	return g.device.SetAutoDetach(autodetach)
}
func (g *gousbDevice) Config(cfgNum int) (Config, error) {
	config, err := g.device.Config(cfgNum)
	if err != nil {
		return nil, err
	}

	return NewGOUSBConfiguration(config), nil
}

func (g *gousbDevice) Descriptor() *gousb.DeviceDesc {
	return g.device.Desc
}

func (g *gousbDevice) Control(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
	return g.device.Control(bmRequestType, bRequest, wValue, wIndex, data)
}

func (g *gousbDevice) Close() error {
	return g.device.Close()
}

func (g *gousbDevice) SerialNumber() (string, error) {
	return g.device.SerialNumber()
}

func (g *gousbDevice) Product() (string, error) {
	return g.device.Product()
}

func (g *gousbDevice) Manufacturer() (string, error) {
	return g.device.Manufacturer()
}

func (g *gousbDevice) GetStringDescriptor(index int) (string, error) {
	return g.device.GetStringDescriptor(index)
}

type gousbContext struct {
	usbCtx *gousb.Context
}

func NewGOUSBContext() Context {
	return &gousbContext{
		usbCtx: gousb.NewContext(),
	}
}

func (g *gousbContext) IterateDevices(reader func(desc *gousb.DeviceDesc)) error {
	_, err := g.usbCtx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		reader(desc)
		return false
	})

	if err != nil {
		return err
	}
	return nil
}

func (g *gousbContext) OpenDevice(vid, pid gousb.ID) (Device, error) {
	device, err := g.usbCtx.OpenDeviceWithVIDPID(vid, pid)
	if err != nil {
		return nil, err
	}

	return NewGOUSBDevice(device)
}

func (g *gousbContext) Close() error {
	return g.usbCtx.Close()
}
