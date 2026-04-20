package mempool

import (
	"bytes"
	"testing"
)

func TestScopeNewWriterBufferCreatesWriter(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(64)

	if w == nil {
		t.Fatal("writer buffer should not be nil")
	}
	if w.Released() {
		t.Fatal("new writer buffer should not be released")
	}
	if w.Len() != 64 {
		t.Fatalf("writer len = %d, want 64", w.Len())
	}
	if w.Cap() < 64 {
		t.Fatalf("writer cap = %d, want >= 64", w.Cap())
	}
	if len(scope.writers) != 1 {
		t.Fatalf("scope writers = %d, want 1", len(scope.writers))
	}
}

func TestWriterBufferToReaderBufferTransfersOwnership(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	w.Reset()
	w.Append([]byte("hello"))

	r := w.ToReaderBuffer()

	if r == nil {
		t.Fatal("reader buffer should not be nil")
	}
	// 验证 writer 所有权已完全转移
	if !w.Released() {
		t.Fatal("writer should be released after transfer")
	}
	if w.buf != nil {
		t.Fatal("writer buf should be nil after transfer")
	}
	if w.pool != nil {
		t.Fatal("writer pool should be nil after transfer")
	}
	if r.Released() {
		t.Fatal("reader should be active after transfer")
	}
	if string(r.Bytes()) != "hello" {
		t.Fatalf("reader bytes = %q, want %q", string(r.Bytes()), "hello")
	}
	if len(scope.readers) != 1 {
		t.Fatalf("scope readers = %d, want 1", len(scope.readers))
	}
}

func TestScopeCloseReleasesTransferredReaderBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))

	r := w.ToReaderBuffer()
	wantCap := r.Cap()

	scope.Close()

	// 验证 reader 已被释放
	if !r.Released() {
		t.Fatal("reader should be released after scope close")
	}
	// 验证缓冲区已归还到池中
	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}
}

func TestScopeCloseReleasesWriterBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	wantCap := w.Cap()

	scope.Close()

	// 验证 writer 已被释放
	if !w.Released() {
		t.Fatal("writer should be released after scope close")
	}
	// 验证缓冲区已归还到池中
	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}
}

func TestWriterBufferToReaderBufferTwicePanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	w.Reset()
	w.Append([]byte("abc"))

	// 第一次转换成功
	_ = w.ToReaderBuffer()

	// 第二次转换应 panic
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on second ToReaderBuffer call")
		}
	}()
	w.ToReaderBuffer()
}

func TestWriterBufferBytesPanicsAfterTransfer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	_ = w.ToReaderBuffer()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on Bytes after transfer")
		}
	}()
	w.Bytes()
}

func TestWriterBufferLenPanicsAfterTransfer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	_ = w.ToReaderBuffer()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on Len after transfer")
		}
	}()
	w.Len()
}

func TestWriterBufferCapPanicsAfterTransfer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	_ = w.ToReaderBuffer()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on Cap after transfer")
		}
	}()
	w.Cap()
}

func TestReaderBufferBytesPanicsAfterScopeClose(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	r := w.ToReaderBuffer()
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on reader Bytes after scope close")
		}
	}()
	r.Bytes()
}

func TestReaderBufferLenPanicsAfterScopeClose(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	r := w.ToReaderBuffer()
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on reader Len after scope close")
		}
	}()
	r.Len()
}

func TestReaderBufferCapPanicsAfterScopeClose(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	r := w.ToReaderBuffer()
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on reader Cap after scope close")
		}
	}()
	r.Cap()
}

func TestWriterBufferAppendGrowsThroughPoolWhenCapacityInsufficient(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(510)
	w.Reset()
	w.Append(bytes.Repeat([]byte{'a'}, 510))

	oldCap := w.Cap()
	w.Append(bytes.Repeat([]byte{'b'}, 10))

	if w.Cap() <= oldCap {
		t.Fatalf("writer cap did not grow, old=%d new=%d", oldCap, w.Cap())
	}
	if w.Len() != 520 {
		t.Fatalf("writer len = %d, want 520", w.Len())
	}
	if w.Cap() < w.Len() {
		t.Fatalf("writer cap = %d, want >= %d", w.Cap(), w.Len())
	}
}

func TestWriterBufferCloneCreatesDetachedCopy(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(10)
	w.Reset()
	w.Append([]byte("hello"))

	dup := w.Clone()
	w.Append([]byte("-world"))

	if string(dup) != "hello" {
		t.Fatalf("clone = %q, want hello", string(dup))
	}
}

func TestReaderBufferCloneCreatesDetachedCopy(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(10)
	w.Reset()
	w.Append([]byte("hello"))
	r := w.ToReaderBuffer()

	dup := r.Clone()
	if string(dup) != "hello" {
		t.Fatalf("clone = %q, want hello", string(dup))
	}
}
