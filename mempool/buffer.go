package mempool

type Buffer struct {
	buf      []byte
	pool     BytePool
	released bool
}

func NewBuffer(pool BytePool, size int) *Buffer {
	return &Buffer{buf: pool.Get(size), pool: pool}
}

func (b *Buffer) Bytes() []byte  { return b.buf }
func (b *Buffer) Len() int       { return len(b.buf) }
func (b *Buffer) Cap() int       { return cap(b.buf) }
func (b *Buffer) Released() bool { return b.released }

func (b *Buffer) Reset() {
	b.mustUsable()
	b.buf = b.buf[:0]
}

func (b *Buffer) Clone() []byte {
	b.mustUsable()
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

func (b *Buffer) EnsureCapacity(additional int) {
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

func (b *Buffer) Resize(n int) {
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

func (b *Buffer) Append(p []byte) {
	b.mustUsable()
	b.EnsureCapacity(len(p))
	start := len(b.buf)
	b.buf = b.buf[:start+len(p)]
	copy(b.buf[start:], p)
}

func (b *Buffer) AppendByte(v byte) {
	b.mustUsable()
	b.EnsureCapacity(1)
	b.buf = append(b.buf, v)
}

func (b *Buffer) DetachedCopy() []byte {
	return b.Clone()
}

func (b *Buffer) Release() {
	b.mustReleasable()
	b.pool.Put(b.buf)
	b.buf = nil
	b.released = true
}
