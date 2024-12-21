package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kanix29/microps/consts"
	"github.com/kanix29/microps/driver"
	"github.com/kanix29/microps/net"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

var terminate = false

func onSignal(sig os.Signal) {
	if sig == syscall.SIGINT {
		terminate = true
	}
}

func main() {
	util.InitLogger()
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		for sig := range sigChan {
			util.Logger.Debug("Signal received in main", zap.String("signal", sig.String()))
			onSignal(sig)
		}
	}()

	// Initialize network
	if err := net.NetInit(); err != nil {
		util.Logger.Error("NetInit() failure", zap.Error(err))
		return
	}

	// Initialize dummy device
	dev, err := driver.DummyInit()
	if err != nil {
		util.Logger.Error("DummyInit() failure", zap.Error(err))
		return
	}

	// Run network
	if err := net.NetRun(); err != nil {
		util.Logger.Error("NetRun() failure", zap.Error(err))
		return
	}

	// Main loop
	for !terminate {
		util.Logger.Debug("Sending data")
		if err := net.NetDeviceOutput(dev, 0x0800, consts.TestData, nil); err != nil {
			util.Logger.Error("NetDeviceOutput() failure", zap.Error(err))
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Shutdown network
	net.NetShutdown()
}
