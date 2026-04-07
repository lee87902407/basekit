package mempool

type HeapBuffer struct {
	buf      []byte
	pool     BytePool
	released bool
}

func NewHeapBuffer(pool BytePool, size int) *HeapBuffer {
	return &HeapBuffer{buf: pool.Get(size), pool: pool}
}

func (b *HeapBuffer) Type() BufferType { return BufferTypeHeap }
func (b *HeapBuffer) Bytes() []byte    { return b.buf }
func (b *HeapBuffer) Len() int         { return len(b.buf) }
func (b *HeapBuffer) Cap() int         { return cap(b.buf) }
func (b *HeapBuffer) Released() bool   { return b.released }

func (b *HeapBuffer) Reset() {
	b.mustUsable()
	b.buf = b.buf[:0]
}

func (b *HeapBuffer) Clone() []byte {
	b.mustUsable()
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

func (b *HeapBuffer) EnsureCapacity(additional int) {
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

func (b *HeapBuffer) Resize(n int) {
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

func (b *HeapBuffer) Append(p []byte) {
	b.mustUsable()
	b.EnsureCapacity(len(p))
	start := len(b.buf)
	b.buf = b.buf[:start+len(p)]
	copy(b.buf[start:], p)
}

func (b *HeapBuffer) AppendByte(v byte) {
	b.mustUsable()
	b.EnsureCapacity(1)
	b.buf = append(b.buf, v)
}

func (b *HeapBuffer) DetachedCopy() []byte {
	return b.Clone()
}

func (b *HeapBuffer) Release() {
	b.mustReleasable()
	b.pool.Put(b.buf)
	b.buf = nil
	b.released = true
}
