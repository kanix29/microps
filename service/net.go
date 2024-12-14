package service

import (
	"fmt"
	"sync"

	"github.com/kanix29/microps/model"
	"github.com/kanix29/microps/util"
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

	fmt.Printf("Registered: dev=%s, type=0x%04x\n", dev.Name, dev.Type)
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
	fmt.Printf("dev=%s, state=%s\n", dev.Name, NET_DEVICE_STATE(dev))
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
	fmt.Printf("dev=%s, state=%s\n", dev.Name, NET_DEVICE_STATE(dev))
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
	fmt.Printf("dev=%s, type=0x%04x, len=%d\n", dev.Name, typ, len(data))
	util.HexDump(data)

	if err := dev.Ops.Transmit(dev, typ, data, dst); err != nil {
		return fmt.Errorf("device transmit failure, dev=%s, len=%d", dev.Name, len(data))
	}
	return nil
}

func NetInputHandler(typ uint16, data []byte, dev *model.NetDevice) error {
	fmt.Printf("dev=%s, type=0x%04x, len=%d\n", dev.Name, typ, len(data))
	util.HexDump(data)
	return nil
}

func NetRun() error {
	fmt.Println("open all devices...")
	for dev := devices; dev != nil; dev = dev.Next {
		if err := NetDeviceOpen(dev); err != nil {
			return err
		}
	}
	fmt.Println("running...")
	return nil
}

func NetShutdown() {
	fmt.Println("close all devices...")
	for dev := devices; dev != nil; dev = dev.Next {
		if err := NetDeviceClose(dev); err != nil {
			fmt.Printf("error closing device: %v\n", err)
		}
	}
	fmt.Println("shutting down")
}

func NetInit() error {
	fmt.Println("initialized")
	return nil
}
