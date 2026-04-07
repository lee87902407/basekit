package mempool

import "unsafe"

type BufferType uint8

const (
	BufferTypeHeap BufferType = iota
	BufferTypeNative
)

type FreeFunc func(unsafe.Pointer)

type BytePool interface {
	Get(size int) []byte
	Put(buf []byte)
	Bucket(size int) int
	MaxPooledCap() int
}

type Buffer interface {
	Type() BufferType
	Bytes() []byte
	Len() int
	Cap() int
	Released() bool
	Reset()
	Clone() []byte
	EnsureCapacity(additional int)
	Resize(n int)
	Append(p []byte)
	AppendByte(v byte)
	DetachedCopy() []byte
	Release()
}

type bufferLifecycle interface {
	Released() bool
	Release()
}

type StatsCollector interface {
	OnGet(size int, bucket int, pooled bool)
	OnPut(capacity int, bucket int, pooled bool)
	OnDrop(capacity int, reason string)
}
