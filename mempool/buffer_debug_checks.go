//go:build debug

package mempool

func (b *Buffer) mustUsable() {
	if b.released {
		panic("mempool: use after release")
	}
}

func (b *Buffer) mustReleasable() {
	if b.released {
		panic("mempool: buffer released twice")
	}
}
