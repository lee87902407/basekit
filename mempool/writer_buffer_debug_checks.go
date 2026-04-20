//go:build debug

package mempool

// mustUsable 检查 WriterBuffer 是否可用，若已释放则 panic。
func (b *WriterBuffer) mustUsable() {
	if b.released {
		panic("mempool: writer buffer is released")
	}
}
