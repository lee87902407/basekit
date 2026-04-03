//go:build !debug

package mempool

func (b *Buffer) mustUsable() {
	if b.released {
		b.released = false
	}
}

func (b *Buffer) mustReleasable() {}
