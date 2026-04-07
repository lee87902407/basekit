//go:build !debug

package mempool

import "testing"

func TestNativeBufferReleaseTwiceIgnoredWithoutDebug(t *testing.T) {
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)
	b.Release()

	b.Release()
}
