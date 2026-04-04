package mempool

type Scope struct {
	pool    BytePool
	buffers []bufferLifecycle
	raws    [][]byte
	closed  bool
}

func NewScope(pool BytePool) *Scope {
	return &Scope{pool: pool}
}

func (s *Scope) Get(size int) []byte {
	buf := s.pool.Get(size)
	s.raws = append(s.raws, buf)
	return buf
}

func (s *Scope) NewBuffer(size int) *Buffer {
	b := NewBuffer(s.pool, size)
	s.buffers = append(s.buffers, b)
	return b
}

func (s *Scope) Track(buf []byte) {
	s.raws = append(s.raws, buf)
}

func (s *Scope) Close() {
	if s.closed {
		return
	}
	for i := range s.buffers {
		if !s.buffers[i].Released() {
			s.buffers[i].Release()
		}
	}
	for i := range s.raws {
		s.pool.Put(s.raws[i])
	}
	s.closed = true
}
