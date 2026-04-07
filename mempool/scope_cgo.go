package mempool

import "C"

import "unsafe"

func (s *Scope) NewNativeBuffer(owner unsafe.Pointer, data *C.char, size int, freeFun FreeFunc) *NativeBuffer {
	if s.closed {
		panic("mempool: scope already closed")
	}
	b := NewNativeBuffer(owner, data, size, freeFun)
	s.buffers = append(s.buffers, b)
	return b
}

func (s *Scope) GetNativeBuffer(owner unsafe.Pointer, data *C.char, size int, freeFun FreeFunc) *NativeBuffer {
	if s.closed {
		panic("mempool: scope already closed")
	}
	b := NewNativeBuffer(owner, data, size, freeFun)
	s.buffers = append(s.buffers, b)
	return b
}
