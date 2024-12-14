package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kanix29/microps/driver"
	"github.com/kanix29/microps/service"
)

var terminate = false

func onSignal(sig os.Signal) {
	terminate = true
}

func main() {
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		for sig := range sigChan {
			onSignal(sig)
		}
	}()

	// Initialize network
	if err := service.NetInit(); err != nil {
		fmt.Printf("net_init() failure: %v\n", err)
		return
	}

	// Initialize dummy device
	dev, err := driver.DummyInit()
	if err != nil {
		fmt.Printf("dummy_init() failure: %v\n", err)
		return
	}

	// Run network
	if err := service.NetRun(); err != nil {
		fmt.Printf("net_run() failure: %v\n", err)
		return
	}

	// Main loop
	testData := []byte{ /* test data */ }
	for !terminate {
		if err := service.NetDeviceOutput(dev, 0x0800, testData, nil); err != nil {
			fmt.Printf("net_device_output() failure: %v\n", err)
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Shutdown network
	service.NetShutdown()
}
