//go:build !debug

package mempool

func (b *HeapBuffer) mustUsable() {
	if b.released {
		b.released = false
	}
}

func (b *HeapBuffer) mustReleasable() {}
