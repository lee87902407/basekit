package mempool

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

type NativeBuffer struct {
	owner    unsafe.Pointer
	freeFun  FreeFunc
	buf      []byte
	released bool
}

func NewNativeBuffer(owner unsafe.Pointer, data *C.char, size int, freeFun FreeFunc) *NativeBuffer {
	if size < 0 {
		panic("mempool: negative native buffer size")
	}
	if size > 0 && data == nil {
		panic("mempool: nil native buffer data")
	}
	if owner != nil && freeFun == nil {
		panic("mempool: nil native buffer free function")
	}

	var buf []byte
	if size > 0 {
		buf = unsafe.Slice((*byte)(unsafe.Pointer(data)), size)
	}

	return &NativeBuffer{owner: owner, freeFun: freeFun, buf: buf}
}

func (b *NativeBuffer) Type() BufferType { return BufferTypeNative }
func (b *NativeBuffer) Bytes() []byte    { return b.buf }
func (b *NativeBuffer) Len() int         { return len(b.buf) }
func (b *NativeBuffer) Cap() int         { return len(b.buf) }
func (b *NativeBuffer) Released() bool   { return b.released }

func (b *NativeBuffer) Reset() {
	panic("mempool: native buffer is read-only")
}

func (b *NativeBuffer) Clone() []byte {
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

func (b *NativeBuffer) EnsureCapacity(additional int) {
	panic("mempool: native buffer is read-only")
}

func (b *NativeBuffer) Resize(n int) {
	panic("mempool: native buffer is read-only")
}

func (b *NativeBuffer) Append(p []byte) {
	panic("mempool: native buffer is read-only")
}

func (b *NativeBuffer) AppendByte(v byte) {
	panic("mempool: native buffer is read-only")
}

func (b *NativeBuffer) DetachedCopy() []byte {
	return b.Clone()
}

func (b *NativeBuffer) Release() {
	b.mustReleasable()
	if b.owner != nil {
		b.freeFun(b.owner)
		b.owner = nil
	}
	b.freeFun = nil
	b.buf = nil
	b.released = true
}
