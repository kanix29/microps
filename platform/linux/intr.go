package platform

import (
	"fmt"
	"os"
	"os/signal"

	"sync"
	"syscall"

	"github.com/kanix29/microps/model"
	"github.com/kanix29/microps/util"
	"go.uber.org/zap"
)

var (
	irqs    *model.IRQEntry
	sigmask = make(map[uint]os.Signal)
	mu      sync.Mutex
	tid     *sync.WaitGroup
	barrier = make(chan struct{})
)

const (
	INTR_IRQ_SHARED = 1      // Define this constant as per your requirement
	INTR_IRQ_BASE   = 34 + 1 // SIGRTMIN + 1
)

func IntrRequestIRQ(irq uint, handler func(irq uint, dev interface{}) error, flags int, name string, dev interface{}) error {
	mu.Lock()
	defer mu.Unlock()

	util.Logger.Debug("RequestIRQ", zap.Uint("irq", irq), zap.Int("flags", flags), zap.String("name", name))
	for entry := irqs; entry != nil; entry = entry.Next {
		if entry.IRQ == irq {
			if (entry.Flags&INTR_IRQ_SHARED == 0) || (flags&INTR_IRQ_SHARED == 0) {
				return fmt.Errorf("conflicts with already registered IRQs")
			}
		}
	}

	entry := &model.IRQEntry{
		IRQ:     irq,
		Handler: handler,
		Flags:   flags,
		Name:    name,
		Dev:     dev,
		Next:    irqs,
	}
	irqs = entry
	sigmask[irq] = syscall.Signal(irq)
	util.Logger.Debug("RequestIRQ registered", zap.Uint("irq", irq), zap.String("name", name))

	return nil
}
func IntrRun() error {
	sigmask[uint(syscall.SIGHUP)] = syscall.SIGHUP
	sigmask[uint(syscall.SIGINT)] = syscall.SIGINT

	// Block signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		for range sigChan {
			// Handle signals if necessary
		}
	}()

	// Create and start the thread
	tid = &sync.WaitGroup{}
	tid.Add(1)
	go func() {
		defer tid.Done()
		IntrThread()
	}()

	// Wait for the barrier
	<-barrier
	return nil
}
func IntrShutdown() {
	if tid == nil {
		// Thread not created
		return
	}
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	tid.Wait()
}

func IntrInit() error {
	tid = &sync.WaitGroup{}
	sigmask[uint(syscall.SIGHUP)] = syscall.SIGHUP
	return nil
}
func IntrThread() {
	terminate := false
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT)

	util.Logger.Info("start...")
	barrier <- struct{}{}

	for !terminate {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT:
			terminate = true
		default:
			for entry := irqs; entry != nil; entry = entry.Next {
				if entry.IRQ == uint(sig.(syscall.Signal)) {
					util.Logger.Debug("IRQ received",
						zap.Uint("irq", entry.IRQ),
						zap.String("name", entry.Name),
					)
					entry.Handler(entry.IRQ, entry.Dev)
				}
			}
		}
	}
	util.Logger.Info("terminated")
}

func IntrRaiseIRQ(irq uint) error {
	if tid == nil {
		return fmt.Errorf("thread not created")
	}
	return syscall.Kill(syscall.Getpid(), syscall.Signal(irq))
}
