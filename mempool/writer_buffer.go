package mempool

// WriterBuffer 是由 Scope 创建的可写缓冲区。
// 写阶段完成后，可通过 ToReaderBuffer 转移所有权到只读的 ReaderBuffer。
type WriterBuffer struct {
	buf      []byte
	pool     BytePool
	scope    *Scope
	released bool
}

// Bytes 返回底层字节切片。
func (b *WriterBuffer) Bytes() []byte {
	b.mustUsable()
	return b.buf
}

// Len 返回缓冲区长度。
func (b *WriterBuffer) Len() int {
	b.mustUsable()
	return len(b.buf)
}

// Cap 返回缓冲区容量。
func (b *WriterBuffer) Cap() int {
	b.mustUsable()
	return cap(b.buf)
}

// Released 返回缓冲区是否已释放。
func (b *WriterBuffer) Released() bool {
	return b.released
}

// ToReaderBuffer 将底层 []byte 所有权转移给 ReaderBuffer。
// 转换后 WriterBuffer 立即失效，不能再继续访问。
func (b *WriterBuffer) ToReaderBuffer() *ReaderBuffer {
	b.mustUsable()
	r := &ReaderBuffer{buf: b.buf, pool: b.pool, released: false}
	b.buf = nil
	b.pool = nil
	b.released = true
	if b.scope != nil {
		b.scope.readers = append(b.scope.readers, r)
	}
	return r
}

// Reset 重置缓冲区长度为零，保留底层数组。
func (b *WriterBuffer) Reset() {
	b.mustUsable()
	b.buf = b.buf[:0]
}

// Clone 返回缓冲区内容的独立副本。
// 返回的切片与原缓冲区不共享内存。
func (b *WriterBuffer) Clone() []byte {
	b.mustUsable()
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

// DetachedCopy 返回缓冲区内容的独立副本，语义与 Clone 一致。
func (b *WriterBuffer) DetachedCopy() []byte {
	return b.Clone()
}

// EnsureCapacity 确保缓冲区至少还有 additional 字节的剩余容量。
// 若容量不足，则通过池分配更大的缓冲区并拷贝已有数据。
func (b *WriterBuffer) EnsureCapacity(additional int) {
	b.mustUsable()
	if additional <= 0 {
		return
	}
	need := len(b.buf) + additional
	if need <= cap(b.buf) {
		return
	}
	next := b.pool.Get(need)
	copy(next, b.buf)
	b.pool.Put(b.buf)
	b.buf = next[:len(b.buf)]
}

// Resize 将缓冲区长度调整为 n。
// 若 n 超过当前容量，则通过池分配更大的缓冲区并拷贝已有数据。
func (b *WriterBuffer) Resize(n int) {
	b.mustUsable()
	if n <= cap(b.buf) {
		b.buf = b.buf[:n]
		return
	}
	next := b.pool.Get(n)
	copy(next, b.buf)
	b.pool.Put(b.buf)
	b.buf = next[:n]
}

// Append 将 p 追加到缓冲区末尾。
// 若容量不足，会自动通过池扩容。
func (b *WriterBuffer) Append(p []byte) {
	b.mustUsable()
	b.EnsureCapacity(len(p))
	start := len(b.buf)
	b.buf = b.buf[:start+len(p)]
	copy(b.buf[start:], p)
}

// AppendByte 将单个字节 v 追加到缓冲区末尾。
// 若容量不足，会自动通过池扩容。
func (b *WriterBuffer) AppendByte(v byte) {
	b.mustUsable()
	b.EnsureCapacity(1)
	b.buf = append(b.buf, v)
}

// releaseToPool 将缓冲区归还到池中。
func (b *WriterBuffer) releaseToPool() {
	if b.released {
		return
	}
	b.pool.Put(b.buf)
	b.buf = nil
	b.released = true
}
