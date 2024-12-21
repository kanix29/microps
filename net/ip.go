package net

import (
	"fmt"

	"github.com/kanix29/microps/model"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

func IpInput(data []byte, dev *model.NetDevice) {
	util.Logger.Debug("IpInput()", zap.String("dev", dev.Name), zap.Int("len", len(data)))
	util.HexDump(data)
}

func IpInit() error {
	if err := NetProtocolRegister(NET_PROTOCOL_TYPE_IP, IpInput); err != nil {
		return fmt.Errorf("NetProtocolRegister() failure: %v", err)
	}
	return nil
}
