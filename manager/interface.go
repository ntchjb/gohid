package manager

import (
	"github.com/google/gousb"
	"github.com/ntchjb/gohid/hid"
)

type DeviceManager interface {
	Init()
	Close() error
	Enumerate(vendorID gousb.ID, productID gousb.ID) (hid.DeviceInfos, error)
	Open(vendorID, productID gousb.ID) (hid.Device, error)
}
