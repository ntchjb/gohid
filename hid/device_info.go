package hid

import (
	"errors"
	"strconv"
	"strings"

	"github.com/google/gousb"
)

var (
	ErrDeviceProfileNotFound = errors.New("device profile not found")
	ErrDeviceDescNotFound    = errors.New("device descriptor not found")
)

type DeviceInfo struct {
	DeviceDesc *gousb.DeviceDesc

	// Connection IDs used by this library
	// 0: Configuration number
	// 1: Interface index of []InterfaceDesc
	// 2: Alternate setting index of []InterfaceSetting
	target [3]int
}

func (d DeviceInfo) String() string {
	var builder strings.Builder
	config := d.DeviceDesc.Configs[d.target[0]]
	intf := config.Interfaces[d.target[1]]
	alt := intf.AltSettings[d.target[2]]
	builder.WriteRune('[')
	builder.WriteString(d.DeviceDesc.Vendor.String())
	builder.WriteRune(':')
	builder.WriteString(d.DeviceDesc.Product.String())
	builder.WriteString("] Conf #")
	builder.WriteString(strconv.Itoa(config.Number))
	builder.WriteString(" Intf #")
	builder.WriteString(strconv.Itoa(intf.Number))
	builder.WriteString(" Sett #")
	builder.WriteString(strconv.Itoa(alt.Alternate))
	builder.WriteString(" Speed: ")
	builder.WriteString(d.DeviceDesc.Speed.String())
	builder.WriteString(", CtrlSize: ")
	builder.WriteString(strconv.Itoa(d.DeviceDesc.MaxControlPacketSize))

	builder.WriteString(", Ep:[")
	i := 0
	for _, endpoint := range alt.Endpoints {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("#" + strconv.Itoa(endpoint.Number))
		if endpoint.Direction {
			builder.WriteString("(IN)")
		} else {
			builder.WriteString("(OUT)")
		}
		i++
	}
	builder.WriteRune(']')

	return builder.String()
}

func (d *DeviceInfo) FromDeviceDesc(desc *gousb.DeviceDesc, confNumber, infNumber, altNumber int) error {
	isFound := false

	for ci, config := range desc.Configs {
		if config.Number != confNumber {
			continue
		}
		for ii, inf := range config.Interfaces {
			if inf.Number != infNumber {
				continue
			}
			for si, setting := range inf.AltSettings {
				if setting.Alternate != altNumber {
					continue
				}
				if setting.Class == gousb.ClassHID {
					d.DeviceDesc = desc
					d.target = [3]int{ci, ii, si}
					return nil
				}
			}
		}
	}

	if !isFound {
		return ErrDeviceProfileNotFound
	}

	return nil
}

func (d *DeviceInfo) GetConfigNumber() int {
	if d.DeviceDesc == nil {
		panic(ErrDeviceDescNotFound)
	}
	return d.DeviceDesc.Configs[d.target[0]].Number
}

func (d *DeviceInfo) GetInterfaceNumber() int {
	if d.DeviceDesc == nil {
		panic(ErrDeviceDescNotFound)
	}
	return d.DeviceDesc.Configs[d.target[0]].Interfaces[d.target[1]].Number
}

func (d *DeviceInfo) GetAltSettingNumber() int {
	if d.DeviceDesc == nil {
		panic(ErrDeviceDescNotFound)
	}
	return d.DeviceDesc.Configs[d.target[0]].Interfaces[d.target[1]].AltSettings[d.target[2]].Alternate
}

func (d *DeviceInfo) GetEndpoints() map[gousb.EndpointAddress]gousb.EndpointDesc {
	if d.DeviceDesc == nil {
		panic(ErrDeviceDescNotFound)
	}
	return d.DeviceDesc.Configs[d.target[0]].Interfaces[d.target[1]].AltSettings[d.target[2]].Endpoints
}

type DeviceInfos []DeviceInfo

func (d DeviceInfos) String() string {
	var res string
	for _, info := range d {
		res += info.String() + "\n"
	}
	return res
}
