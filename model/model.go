package model

const (
	// IFNAMSIZ: インターフェース名の最大サイズ
	IFNAMSIZ = 16

	// ネットワークデバイスのタイプ
	NET_DEVICE_TYPE_DUMMY    = 0x0000
	NET_DEVICE_TYPE_LOOPBACK = 0x0001
	NET_DEVICE_TYPE_ETHERNET = 0x0002

	// ネットワークデバイスのフラグ
	NET_DEVICE_FLAG_UP        = 0x0001
	NET_DEVICE_FLAG_LOOPBACK  = 0x0010
	NET_DEVICE_FLAG_BROADCAST = 0x0020
	NET_DEVICE_FLAG_P2P       = 0x0040
	NET_DEVICE_FLAG_NEED_ARP  = 0x0100

	// ネットワークデバイスアドレスの長さ
	NET_DEVICE_ADDR_LEN = 16

	// ダミーデバイスのMTU
	DUMMY_MTU = ^uint16(0) // UINT16_MAX
)

type NetDevice struct {
	Next      *NetDevice
	Index     uint
	Name      string
	Type      uint16
	MTU       uint16
	Flags     uint16
	Hlen      uint16 // Header length
	Alen      uint16 // Address length
	Addr      [NET_DEVICE_ADDR_LEN]byte
	Peer      [NET_DEVICE_ADDR_LEN]byte
	Broadcast [NET_DEVICE_ADDR_LEN]byte
	Ops       *NetDeviceOps
	Priv      interface{} // 汎用ポインタとして使用
}

// ネットワークデバイス操作の構造体
type NetDeviceOps struct {
	Open     func(dev *NetDevice) error
	Close    func(dev *NetDevice) error
	Transmit func(dev *NetDevice, t uint16, data []byte, dst interface{}) error
}

type IRQEntry struct {
	Next    *IRQEntry
	IRQ     uint
	Handler func(irq uint, dev interface{}) error
	Flags   int
	Name    string
	Dev     interface{}
}
