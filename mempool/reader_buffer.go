package mempool

// ReaderBuffer 是只读缓冲区
type ReaderBuffer struct {
	buf []byte
	idx int
	cap int
}

// Len 返回缓冲区长度。
func (b *ReaderBuffer) Len() int {
	return len(b.buf)
}

// Cap 返回缓冲区容量。
func (b *ReaderBuffer) Cap() int {
	return cap(b.buf)
}
