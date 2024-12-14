package util

import (
	"testing"
)

func TestLprintf(t *testing.T) {
	Lprintf('I', "util_test.go", 10, "TestLprintf", "This is a test message: %d", 123)
}

func TestHexDump(t *testing.T) {
	data := []byte("This is a test for hex dump function.")
	HexDump(data)
}

func TestQueue(t *testing.T) {
	var queue QueueHead
	QueueInit(&queue)
	QueuePush(&queue, "first")
	QueuePush(&queue, "second")
	if QueuePop(&queue) != "first" {
		t.Error("Expected 'first'")
	}
	if QueuePeek(&queue) != "second" {
		t.Error("Expected 'second'")
	}
	if QueuePop(&queue) != "second" {
		t.Error("Expected 'second'")
	}
	if QueuePop(&queue) != nil {
		t.Error("Expected nil")
	}
}

func TestByteorder(t *testing.T) {
	h := uint16(0x1234)
	n := Hton16(h)
	if n != 0x3412 {
		t.Errorf("Expected 0x3412, got 0x%04x", n)
	}
	h = Ntoh16(n)
	if h != 0x1234 {
		t.Errorf("Expected 0x1234, got 0x%04x", h)
	}

	h32 := uint32(0x12345678)
	n32 := Hton32(h32)
	if n32 != 0x78563412 {
		t.Errorf("Expected 0x78563412, got 0x%08x", n32)
	}
	h32 = Ntoh32(n32)
	if h32 != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%08x", h32)
	}
}

func TestCksum16(t *testing.T) {
	data := []uint16{0x1234, 0x5678, 0x9abc, 0xdef0}
	sum := Checksum16(data, 0)
	if sum != 0x1e1e {
		t.Errorf("Expected 0x1e1e, got 0x%04x", sum)
	}
}
