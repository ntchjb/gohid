package manager

import (
	"github.com/ntchjb/gohid/hid"
)

type DeviceManager interface {
	Init()
	Close() error
	Enumerate(vendorID uint16, productID uint16) (hid.DeviceInfos, error)
	Open(deviceInfo hid.DeviceInfo) (hid.DeviceAccessor, error)
}
