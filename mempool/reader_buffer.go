package mempool

// ReaderBuffer 是只读缓冲区，由 WriterBuffer.ToReaderBuffer() 产生。
type ReaderBuffer struct {
	buf      []byte
	pool     BytePool
	released bool
}

// Bytes 返回底层字节切片。
func (b *ReaderBuffer) Bytes() []byte {
	b.mustUsable()
	return b.buf
}

// Len 返回缓冲区长度。
func (b *ReaderBuffer) Len() int {
	b.mustUsable()
	return len(b.buf)
}

// Cap 返回缓冲区容量。
func (b *ReaderBuffer) Cap() int {
	b.mustUsable()
	return cap(b.buf)
}

// Released 返回缓冲区是否已释放。
func (b *ReaderBuffer) Released() bool {
	return b.released
}

// Clone 返回缓冲区内容的独立副本。
// 返回的切片与原缓冲区不共享内存。
func (b *ReaderBuffer) Clone() []byte {
	b.mustUsable()
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

// DetachedCopy 返回缓冲区内容的独立副本，语义与 Clone 一致。
func (b *ReaderBuffer) DetachedCopy() []byte {
	return b.Clone()
}

// releaseToPool 将缓冲区归还到池中。
func (b *ReaderBuffer) releaseToPool() {
	if b.released {
		return
	}
	b.pool.Put(b.buf)
	b.buf = nil
	b.released = true
}
