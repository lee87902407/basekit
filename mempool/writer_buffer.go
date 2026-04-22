package mempool

// WriterBuffer 是由 Scope 创建的可写缓冲区。
type WriterBuffer struct {
	buf      []byte
	idx      int
	capacity int
	scope    *Scope
}

// Len 返回缓冲区长度。
func (b *WriterBuffer) Len() int {
	return b.idx
}

// Cap 返回缓冲区容量。
func (b *WriterBuffer) Cap() int {
	return b.capacity
}

// Reset 重置缓冲区长度为零，保留底层数组。
func (b *WriterBuffer) Reset() {
	b.buf = b.buf[:0]
	b.idx = 0
}

func (b *WriterBuffer) CloneByBuffer() *WriterBuffer {
	dup := b.scope.NewWriterBuffer(b.capacity)
	dup.idx = b.idx
	dup.capacity = b.capacity
	copy(dup.buf, b.buf)
	return dup
}

//// EnsureCapacity 确保缓冲区至少还有 additional 字节的剩余容量。
//// 若容量不足，则通过池分配更大的缓冲区并拷贝已有数据。
//func (b *WriterBuffer) ensureCapacity(additional int) {
//	if additional <= 0 {
//		return
//	}
//	need := len(b.buf) + additional
//	if need <= cap(b.buf) {
//		return
//	}
//	next := b.scope.Get(need)
//	copy(next, b.buf)
//	b.buf = next[:len(b.buf)]
//}

//// Resize 将缓冲区长度调整为 n。
//// 若 n 超过当前容量，则通过池分配更大的缓冲区并拷贝已有数据。
//func (b *WriterBuffer) Resize(n int) {
//	b.mustUsable()
//	if n <= cap(b.buf) {
//		b.buf = b.buf[:n]
//		return
//	}
//	next := b.scope.Get(n)
//	copy(next, b.buf)
//	b.buf = next[:n]
//}

// Append 将 p 追加到缓冲区末尾。若容量不足,直接panic
func (b *WriterBuffer) Append(p []byte) {
	//b.mustUsable()
	//b.ensureCapacity(len(p))
	if len(p) > b.capacity-b.idx {
		panic("mempool: buffer overflow")
	}

	b.buf = b.buf[:b.idx+len(p)]
	copy(b.buf[b.idx:], p)
	b.idx += len(p)
}

// AppendByte 将单个字节 v 追加到缓冲区末尾。
// 若容量不足,直接panic
func (b *WriterBuffer) AppendByte(v byte) {
	if b.capacity-b.idx < 1 {
		panic("mempool: buffer overflow")
	}
	b.buf = append(b.buf, v)
	b.idx++
}
