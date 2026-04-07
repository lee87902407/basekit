package mempool

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

func newOwnedNativeBufferForTesting(data []byte) (unsafe.Pointer, *C.char, int) {
	size := len(data)
	owner := C.malloc(C.size_t(size))
	if owner == nil && size > 0 {
		panic("mempool: malloc failed in test helper")
	}
	if size > 0 {
		buf := unsafe.Slice((*byte)(owner), size)
		copy(buf, data)
	}
	return owner, (*C.char)(owner), size
}

func freeOwnedNativeBufferForTesting(owner unsafe.Pointer) {
	if owner != nil {
		C.free(owner)
	}
}

func nativeBufferTestFreeFunc(owner unsafe.Pointer) {
	freeOwnedNativeBufferForTesting(owner)
}
