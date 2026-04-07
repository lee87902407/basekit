//go:build debug

package mempool

func (b *NativeBuffer) mustReleasable() {
	if b.released {
		panic("mempool: buffer released twice")
	}
}
