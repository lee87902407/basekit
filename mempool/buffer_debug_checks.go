//go:build debug

package mempool

func (b *HeapBuffer) mustUsable() {
	if b.released {
		panic("mempool: use after release")
	}
}

func (b *HeapBuffer) mustReleasable() {
	if b.released {
		panic("mempool: buffer released twice")
	}
}
