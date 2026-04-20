//go:build debug

package mempool

// mustUsable 检查 ReaderBuffer 是否可用，若已释放则 panic。
func (b *ReaderBuffer) mustUsable() {
	if b.released {
		panic("mempool: reader buffer is released")
	}
}
