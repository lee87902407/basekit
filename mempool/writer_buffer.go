package mempool

// WriterBuffer 是由 Scope 创建的可写缓冲区。
type WriterBuffer struct {
	buf      []byte
	idx      int
	capacity int
}

// WriteBytes 返回还能写入的大小
func (b *WriterBuffer) WriteBytes() int {
	return b.capacity - b.idx
}

// Len 返回当前已写入长度。
func (b *WriterBuffer) Len() int {
	return b.idx
}

// Cap 返回缓冲区容量。
func (b *WriterBuffer) Cap() int {
	return b.capacity
}

// Reset 重置写入长度为零，保留底层数组。
func (b *WriterBuffer) Reset() {
	b.idx = 0
}

// Append 将 p 追加到缓冲区末尾。若容量不足,直接panic
func (b *WriterBuffer) Append(p []byte) {
	if len(p) > b.capacity-b.idx {
		panic("mempool: buffer overflow")
	}
	copy(b.buf[b.idx:], p)
	b.idx += len(p)
}

// AppendByte 将单个字节 v 追加到缓冲区末尾。
// 若容量不足,直接panic
func (b *WriterBuffer) AppendByte(v byte) {
	if b.capacity-b.idx < 1 {
		panic("mempool: buffer overflow")
	}
	b.buf[b.idx] = v
	b.idx++
}
