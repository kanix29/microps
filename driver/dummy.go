package driver

import (
	"fmt"

	"github.com/kanix29/microps/model"
	"github.com/kanix29/microps/service"
	"github.com/kanix29/microps/util"
)

func DummyTransmit(dev *model.NetDevice, typ uint16, data []byte, dst interface{}) error {
	fmt.Printf("dev=%s, type=0x%04x, len=%d\n", dev.Name, typ, len(data))
	util.HexDump(data)
	// Drop data
	return nil
}

var DummyOps = &model.NetDeviceOps{
	Transmit: DummyTransmit,
}

func DummyInit() (*model.NetDevice, error) {
	dev := service.NetDeviceAlloc()
	if dev == nil {
		return nil, fmt.Errorf("net_device_alloc() failure")
	}
	dev.Type = model.NET_DEVICE_TYPE_DUMMY
	dev.MTU = model.DUMMY_MTU
	dev.Hlen = 0 // non header
	dev.Alen = 0 // non address
	dev.Ops = DummyOps
	if err := service.NetDeviceRegister(dev); err != nil {
		return nil, fmt.Errorf("net_device_register() failure")
	}
	fmt.Printf("initialized, dev=%s\n", dev.Name)
	return dev, nil
}
