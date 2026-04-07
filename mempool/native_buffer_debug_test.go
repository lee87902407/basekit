//go:build debug

package mempool

import "testing"

func TestNativeBufferReleaseTwicePanics(t *testing.T) {
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)
	b.Release()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on double release")
		}
	}()

	b.Release()
}
