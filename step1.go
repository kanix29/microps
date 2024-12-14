package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kanix29/microps/driver"
	"github.com/kanix29/microps/service"
	"github.com/kanix29/microps/test"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

var terminate = false

func onSignal(sig os.Signal) {
	terminate = true
}

func main() {
	util.InitLogger()
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		for sig := range sigChan {
			onSignal(sig)
		}
	}()

	// Initialize network
	if err := service.NetInit(); err != nil {
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
	if err := service.NetRun(); err != nil {
		util.Logger.Error("NetRun() failure", zap.Error(err))
		return
	}

	// Main loop
	for !terminate {
		if err := service.NetDeviceOutput(dev, 0x0800, test.TestData, nil); err != nil {
			util.Logger.Error("NetDeviceOutput() failure", zap.Error(err))
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Shutdown network
	service.NetShutdown()
}
