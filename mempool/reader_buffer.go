package mempool

// ReaderBuffer 是固定大小的只读缓冲区视图。
// 不提供修改有效区间或暴露底层切片的方法。
type ReaderBuffer struct {
	buf []byte
	idx int
	cap int
}

// Len 返回缓冲区有效长度。
func (b *ReaderBuffer) Len() int {
	return b.cap
}

// Cap 返回缓冲区容量（与 Len 相同，因为 ReaderBuffer 是固定大小的只读视图）。
func (b *ReaderBuffer) Cap() int {
	return b.cap
}

// ByteAt 返回有效区间内第 i 个字节的只读访问。
// 若 i 越界则 panic。
func (b *ReaderBuffer) ByteAt(i int) byte {
	if i < 0 || i >= b.cap {
		panic("mempool: ReaderBuffer index out of range")
	}
	return b.buf[i]
}
