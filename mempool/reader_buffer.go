package mempool

// ReaderBuffer 是只读缓冲区
type ReaderBuffer struct {
	buf []byte
	idx int
	cap int
}

// Len 返回缓冲区长度。
func (b *ReaderBuffer) Len() int {
	return b.cap
}

func (b *ReaderBuffer) IndexByte(n int) byte {
	return b.buf[n]
}

func (b *ReaderBuffer) RemoveLast(n int) {
	b.cap -= 2
}
func (b *ReaderBuffer) Slice(start, end int) []byte {
	return b.buf[start:end]
}
