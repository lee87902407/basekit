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

func (s *Scope) NewWriterBuffer(capacity int) *WriterBuffer {
	b := &WriterBuffer{buf: s.pool.get(capacity), scope: s, idx: 0, capacity: capacity}
	s.writers = append(s.writers, b)
	return b
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
