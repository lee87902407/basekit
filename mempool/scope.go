package mempool

type Scope struct {
	pool    BytePool
	writers []*WriterBuffer
	readers []*ReaderBuffer
	raws    [][]byte
	closed  bool
}

func NewScope(pool BytePool) *Scope {
	return &Scope{pool: pool}
}

// mustOpen 检查 Scope 是否已关闭，若已关闭则 panic。
func (s *Scope) mustOpen() {
	if s.closed {
		panic("mempool: scope is closed")
	}
}

func (s *Scope) Get(size int) []byte {
	s.mustOpen()
	buf := s.pool.Get(size)
	s.raws = append(s.raws, buf)
	return buf
}

// NewWriterBuffer 创建一个 WriterBuffer 并纳入 Scope 管理。
func (s *Scope) NewWriterBuffer(size int) *WriterBuffer {
	s.mustOpen()
	b := &WriterBuffer{buf: s.pool.Get(size), pool: s.pool, scope: s}
	s.writers = append(s.writers, b)
	return b
}

func (s *Scope) Track(buf []byte) {
	s.mustOpen()
	s.raws = append(s.raws, buf)
}

func (s *Scope) Close() {
	if s.closed {
		return
	}
	for i := range s.writers {
		s.writers[i].releaseToPool()
	}
	for i := range s.readers {
		s.readers[i].releaseToPool()
	}
	for i := range s.raws {
		s.pool.Put(s.raws[i])
	}
	s.closed = true
}
