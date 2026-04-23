package mempool

type Scope struct {
	pool    *BytePool
	writers []*WriterBuffer
	readers []*ReaderBuffer
	raws    [][]byte
}

func (s *Scope) Get(size int) []byte {
	buf := s.pool.get(size)
	s.raws = append(s.raws, buf)
	return buf
}

func (b *Scope) CloneByBuffer(buf *WriterBuffer) *WriterBuffer {
	dup := b.NewWriterBuffer(buf.capacity)
	copy(dup.buf, buf.buf)
	dup.idx = buf.idx
	dup.capacity = buf.capacity
	return dup
}

func (b *Scope) CloneFromBytes(bytes []byte) *WriterBuffer {
	var capacity = cap(bytes)
	var buf = b.NewWriterBuffer(capacity)
	buf.Append(bytes)
	return buf
}

func (s *Scope) NewWriterBuffer(capacity int) *WriterBuffer {
	b := &WriterBuffer{buf: s.pool.get(capacity), idx: 0, capacity: capacity}
	s.writers = append(s.writers, b)
	return b
}

func (s *Scope) ResetWriterBufferByCapacity(buf *WriterBuffer, capacity int) {
	if capacity <= 0 {
		panic("invalid capacity")
	}
	if capacity < buf.idx {
		buf.idx = capacity
		buf.capacity = capacity
		return
	}
	if capacity <= cap(buf.buf) {
		buf.capacity = capacity
		return
	}

	next := s.pool.get(capacity)
	copy(next[:buf.idx], buf.buf[:buf.idx])
	s.pool.put(buf.buf)
	buf.buf = next
	buf.capacity = capacity
}

func (s *Scope) Close() {

	if s.pool == nil {
		panic("mempool pool is nil,scope is already closed")
	}

	for i := range s.writers {
		if s.writers[i].buf != nil {
			s.pool.put(s.writers[i].buf)
			s.writers[i].buf = nil
		}
	}

	for i := range s.readers {
		if s.readers[i].buf != nil {
			s.pool.put(s.readers[i].buf)
			s.readers[i].buf = nil
		}
	}

	for i := range s.raws {
		s.pool.put(s.raws[i])
	}

	s.raws = nil

	s.pool = nil
}

func (s *Scope) ToReaderBuffer(w *WriterBuffer) *ReaderBuffer {
	r := &ReaderBuffer{buf: w.buf, idx: 0, cap: w.idx}
	w.buf = nil
	s.readers = append(s.readers, r)
	return r
}
