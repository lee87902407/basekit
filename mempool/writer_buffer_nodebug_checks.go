//go:build !debug

package mempool

// mustUsable 检查 WriterBuffer 是否可用，若已释放则 panic。
// 与 debug 版本行为一致：释放后的对象不可用。
func (b *WriterBuffer) mustUsable() {
	if b.released {
		panic("mempool: writer buffer is released")
	}
}
