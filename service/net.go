package service

import (
	"fmt"
	"sync"

	"github.com/kanix29/microps/model"
	platform "github.com/kanix29/microps/platform/linux"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

func NET_DEVICE_IS_UP(dev *model.NetDevice) bool {
	return dev.Flags&model.NET_DEVICE_FLAG_UP != 0
}

func NET_DEVICE_STATE(dev *model.NetDevice) string {
	if NET_DEVICE_IS_UP(dev) {
		return "up"
	}
	return "down"
}

// Global list of devices and mutex for thread safety.
var (
	devices     *model.NetDevice
	deviceMutex sync.Mutex
	deviceIndex uint
)

// NetDeviceAlloc allocates a new NetDevice.
func NetDeviceAlloc() *model.NetDevice {
	dev := &model.NetDevice{}
	return dev
}

// NetDeviceRegister registers a network device.
// NOTE: Must not be called after netRun().
func NetDeviceRegister(dev *model.NetDevice) error {
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	dev.Index = deviceIndex
	dev.Name = fmt.Sprintf("net%d", dev.Index)
	dev.Next = devices
	devices = dev

	util.Logger.Debug("NetDeviceRegister", zap.String("dev", dev.Name), zap.String("type", fmt.Sprintf("0x%04x", dev.Type)))
	deviceIndex++

	return nil
}

func NetDeviceOpen(dev *model.NetDevice) error {
	if NET_DEVICE_IS_UP(dev) {
		return fmt.Errorf("already opened, dev=%s", dev.Name)
	}
	if dev.Ops.Open != nil {
		if err := dev.Ops.Open(dev); err != nil {
			return fmt.Errorf("failure, dev=%s", dev.Name)
		}
	}
	dev.Flags |= model.NET_DEVICE_FLAG_UP
	util.Logger.Debug("NetDeviceOpen", zap.String("dev", dev.Name), zap.String("state", NET_DEVICE_STATE(dev)))
	return nil
}

func NetDeviceClose(dev *model.NetDevice) error {
	if !NET_DEVICE_IS_UP(dev) {
		return fmt.Errorf("not opened, dev=%s", dev.Name)
	}
	if dev.Ops.Close != nil {
		if err := dev.Ops.Close(dev); err != nil {
			return fmt.Errorf("failure, dev=%s", dev.Name)
		}
	}
	dev.Flags &^= model.NET_DEVICE_FLAG_UP
	util.Logger.Debug("NetDeviceClose", zap.String("dev", dev.Name), zap.String("state", NET_DEVICE_STATE(dev)))
	return nil
}

func NetDeviceOutput(dev *model.NetDevice, typ uint16, data []byte, dst interface{}) error {

	// Configure the device to be up
	if !NET_DEVICE_IS_UP(dev) {
		return fmt.Errorf("not opened, dev=%s", dev.Name)
	}

	// Check if the data is too long
	if len(data) > int(dev.MTU) {
		return fmt.Errorf("too long, dev=%s, mtu=%d, len=%d", dev.Name, dev.MTU, len(data))
	}

	// Use HexDump function to print the data in a debug format
	util.Logger.Debug("NetDeviceOutput", zap.String("dev", dev.Name), zap.String("type", fmt.Sprintf("0x%04x", typ)), zap.Int("len", len(data)))
	util.HexDump(data)

	if err := dev.Ops.Transmit(dev, typ, data, dst); err != nil {
		return fmt.Errorf("device transmit failure, dev=%s, len=%d", dev.Name, len(data))
	}
	return nil
}

func NetInputHandler(typ uint16, data []byte, dev *model.NetDevice) error {
	util.Logger.Debug("NetInputHandler", zap.String("dev", dev.Name), zap.String("type", fmt.Sprintf("0x%04x", typ)), zap.Int("len", len(data)))
	util.HexDump(data)
	return nil
}

func NetRun() error {
	if err := platform.IntrRun(); err != nil {
		return fmt.Errorf("intr_run() failure")
	}
	util.Logger.Info("NetRun: opened all devices...")
	for dev := devices; dev != nil; dev = dev.Next {
		if err := NetDeviceOpen(dev); err != nil {
			return err
		}
	}
	util.Logger.Info("NetRun: running...")
	return nil
}

func NetShutdown() {
	util.Logger.Info("NetShutdown: closing all devices...")
	for dev := devices; dev != nil; dev = dev.Next {
		if err := NetDeviceClose(dev); err != nil {
			util.Logger.Error("NetShutdown: error closing device", zap.Error(err))
		}
	}
	platform.IntrShutdown()
	util.Logger.Info("NetShutdown: shutting down")
}

func NetInit() error {
	if err := platform.IntrInit(); err != nil {
		return fmt.Errorf("intr_init() failure")
	}
	util.Logger.Info("NetInit: initialized")
	return nil
}
