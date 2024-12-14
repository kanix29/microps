package util

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"unsafe"
)

/*
 * Logging
 */

func Lprintf(level byte, file string, line int, function string, format string, a ...interface{}) (n int, err error) {
	now := time.Now()
	timestamp := now.Format("15:04:05.000")
	msg := fmt.Sprintf(format, a...)
	n, err = fmt.Fprintf(os.Stderr, "%s [%c] %s: %s (%s:%d)\n", timestamp, level, function, msg, file, line)
	return
}

func HexDump(data []byte) {
	fmt.Fprintf(os.Stderr, "+------+-------------------------------------------------+------------------+\n")
	for offset := 0; offset < len(data); offset += 16 {
		fmt.Fprintf(os.Stderr, "| %04x | ", offset)
		for index := 0; index < 16; index++ {
			if offset+index < len(data) {
				fmt.Fprintf(os.Stderr, "%02x ", data[offset+index])
			} else {
				fmt.Fprintf(os.Stderr, "   ")
			}
		}
		fmt.Fprintf(os.Stderr, "| ")
		for index := 0; index < 16; index++ {
			if offset+index < len(data) {
				if data[offset+index] >= 32 && data[offset+index] <= 126 {
					fmt.Fprintf(os.Stderr, "%c", data[offset+index])
				} else {
					fmt.Fprintf(os.Stderr, ".")
				}
			} else {
				fmt.Fprintf(os.Stderr, " ")
			}
		}
		fmt.Fprintf(os.Stderr, " |\n")
	}
	fmt.Fprintf(os.Stderr, "+------+-------------------------------------------------+------------------+\n")
}

/*
 * Queue
 */

type QueueEntry struct {
	Data interface{}
	Next *QueueEntry
}

type QueueHead struct {
	Head *QueueEntry
	Tail *QueueEntry
	Num  uint
}

func QueueInit(queue *QueueHead) {
	queue.Head = nil
	queue.Tail = nil
	queue.Num = 0
}

func QueuePush(queue *QueueHead, data interface{}) {
	entry := &QueueEntry{Data: data}
	if queue.Tail != nil {
		queue.Tail.Next = entry
	} else {
		queue.Head = entry
	}
	queue.Tail = entry
	queue.Num++
}

func QueuePop(queue *QueueHead) interface{} {
	if queue.Head == nil {
		return nil
	}
	entry := queue.Head
	queue.Head = entry.Next
	if queue.Head == nil {
		queue.Tail = nil
	}
	queue.Num--
	return entry.Data
}

func QueuePeek(queue *QueueHead) interface{} {
	if queue.Head == nil {
		return nil
	}
	return queue.Head.Data
}

func QueueForeach(queue *QueueHead, fn func(arg interface{}, data interface{}), arg interface{}) {
	for entry := queue.Head; entry != nil; entry = entry.Next {
		fn(arg, entry.Data)
	}
}

/*
 * ByteOrder
 */

// Hton16 converts host byte order to network byte order for 16-bit value.
func Hton16(h uint16) uint16 {
	return binary.BigEndian.Uint16((*[2]byte)(unsafe.Pointer(&h))[:])
}

// Ntoh16 converts network byte order to host byte order for 16-bit value.
func Ntoh16(n uint16) uint16 {
	return binary.BigEndian.Uint16((*[2]byte)(unsafe.Pointer(&n))[:])
}

// Hton32 converts host byte order to network byte order for 32-bit value.
func Hton32(h uint32) uint32 {
	return binary.BigEndian.Uint32((*[4]byte)(unsafe.Pointer(&h))[:])
}

// Ntoh32 converts network byte order to host byte order for 32-bit value.
func Ntoh32(n uint32) uint32 {
	return binary.BigEndian.Uint32((*[4]byte)(unsafe.Pointer(&n))[:])
}

/*
 * Checksum
 */

func Checksum16(addr []uint16, init uint32) uint16 {
	var sum uint32 = init
	for _, v := range addr {
		sum += uint32(v)
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	return uint16(^sum)
}
