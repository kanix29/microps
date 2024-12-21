package driver

import (
	"fmt"
	"sync"

	"github.com/kanix29/microps/model"
	"github.com/kanix29/microps/net"
	platform "github.com/kanix29/microps/platform/linux"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

const (
	LOOPBACK_MTU         = ^uint16(0)
	LOOPBACK_QUEUE_LIMIT = 16
	LOOPBACK_IRQ         = platform.INTR_IRQ_BASE + 1
)

// LoopbackFromPriv は、NetDeviceのPrivフィールドからLoopback構造体を取得します。
func LoopbackFromPriv(dev *model.NetDevice) *Loopback {
	if lo, ok := dev.Priv.(*Loopback); ok {
		return lo
	}
	return nil
}

type Loopback struct {
	irq   uint
	mutex sync.Mutex
	queue util.QueueHead
}

type LoopbackQueueEntry struct {
	Type uint16
	Len  int
	Data []byte
}

var LoopbackOps = &model.NetDeviceOps{
	Transmit: LoopbackTransmit,
}

// func DummyISR(irq uint, id interface{}) error {
// 	dev, ok := id.(*model.NetDevice)
// 	if !ok {
// 		return fmt.Errorf("invalid device id\n")
// 	}
// 	util.Logger.Debug("DummyISR", zap.Uint("irq", irq), zap.String("dev", dev.Name))
// 	return nil
// }

func LoopbackTransmit(dev *model.NetDevice, typ uint16, data []byte, dst interface{}) error {
	lo := LoopbackFromPriv(dev)
	if lo == nil {
		return fmt.Errorf("invalid device")
	}

	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	if lo.queue.Num >= LOOPBACK_QUEUE_LIMIT {
		return fmt.Errorf("queue is full")
	}

	entry := &LoopbackQueueEntry{
		Type: typ,
		Len:  len(data),
		Data: make([]byte, len(data)),
	}
	copy(entry.Data, data)
	util.QueuePush(&lo.queue, entry)
	num := lo.queue.Num

	util.Logger.Debug("LoopbackTransmit() queue pushed", zap.Uint("num", num), zap.String("dev", dev.Name), zap.String("type", fmt.Sprintf("0x%04x", typ)), zap.Int("len", len(data)))
	util.HexDump(data)

	platform.IntrRaiseIRQ(lo.irq)
	return nil
}

func LoopbackISR(irq uint, id interface{}) error {
	dev, ok := id.(*model.NetDevice)
	if !ok {
		return fmt.Errorf("invalid device id")
	}

	lo := LoopbackFromPriv(dev)
	if lo == nil {
		return fmt.Errorf("invalid device")
	}

	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	for {
		entryInterface := util.QueuePop(&lo.queue)
		if entryInterface == nil {
			break
		}
		entry, ok := entryInterface.(*LoopbackQueueEntry)
		if !ok {
			return fmt.Errorf("invalid queue entry type")
		}
		util.Logger.Debug("LoopbackISR() queue popped", zap.Uint("num", lo.queue.Num), zap.String("dev", dev.Name), zap.String("type", fmt.Sprintf("0x%04x", entry.Type)), zap.Int("len", entry.Len))
		util.HexDump(entry.Data)

		net.NetInputHandler(entry.Type, entry.Data, dev)
		// golangでは、ガベージコレクションによってメモリ管理されているので不要
		// memory_free(entry);
	}

	return nil
}

func LoopbackInit() (*model.NetDevice, error) {
	// exercise3-1
	dev := net.NetDeviceAlloc()
	if dev == nil {
		return nil, fmt.Errorf("NetDeviceAlloc() failure")
	}
	dev.Type = model.NET_DEVICE_TYPE_LOOPBACK
	dev.MTU = LOOPBACK_MTU
	dev.Hlen = 0 // non header
	dev.Alen = 0 // non address
	dev.Flags = model.NET_DEVICE_FLAG_LOOPBACK
	dev.Ops = LoopbackOps

	lo := &Loopback{
		irq: LOOPBACK_IRQ,
	}
	util.QueueInit(&lo.queue)
	dev.Priv = lo

	// exercise3-2
	if err := net.NetDeviceRegister(dev); err != nil {
		return nil, fmt.Errorf("NetDeviceRegister() failure")
	}

	if err := platform.IntrRequestIRQ(LOOPBACK_IRQ, LoopbackISR, platform.INTR_IRQ_SHARED, dev.Name, dev); err != nil {
		return nil, fmt.Errorf("IntrRequestIRQ() failure: %v", err)
	}

	util.Logger.Debug("LoopbackInit: initialized", zap.String("dev", dev.Name))
	return dev, nil
}
