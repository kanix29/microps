package driver

import (
	"fmt"

	"github.com/kanix29/microps/model"
	"github.com/kanix29/microps/net"
	platform "github.com/kanix29/microps/platform/linux"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

const DUMMY_IRQ = platform.INTR_IRQ_BASE

var DummyOps = &model.NetDeviceOps{
	Transmit: DummyTransmit,
}

func DummyTransmit(dev *model.NetDevice, typ uint16, data []byte, dst interface{}) error {
	util.Logger.Debug("DummyTransmit", zap.String("dev", dev.Name), zap.String("type", fmt.Sprintf("0x%04x", typ)), zap.Int("len", len(data)))
	util.HexDump(data)
	// Drop data
	platform.IntrRaiseIRQ(DUMMY_IRQ)
	return nil
}

func DummyISR(irq uint, id interface{}) error {
	dev, ok := id.(*model.NetDevice)
	if !ok {
		return fmt.Errorf("invalid device id\n")
	}
	util.Logger.Debug("DummyISR", zap.Uint("irq", irq), zap.String("dev", dev.Name))
	return nil
}

func DummyInit() (*model.NetDevice, error) {
	dev := net.NetDeviceAlloc()
	if dev == nil {
		return nil, fmt.Errorf("net_device_alloc() failure")
	}
	dev.Type = model.NET_DEVICE_TYPE_DUMMY
	dev.MTU = model.DUMMY_MTU
	dev.Hlen = 0 // non header
	dev.Alen = 0 // non address
	dev.Ops = DummyOps
	if err := net.NetDeviceRegister(dev); err != nil {
		return nil, fmt.Errorf("net_device_register() failure")
	}
	if err := platform.IntrRequestIRQ(DUMMY_IRQ, DummyISR, platform.INTR_IRQ_SHARED, dev.Name, dev); err != nil {
		return nil, fmt.Errorf("intr_request_irq() failure: %v", err)
	}
	util.Logger.Debug("DummyInit: initialized", zap.String("dev", dev.Name))
	return dev, nil
}
