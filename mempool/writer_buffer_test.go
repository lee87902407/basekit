package mempool

import "testing"

func TestScopeResetWriterBufferByCapacityTruncatesWrittenData(t *testing.T) {
	pool := New(DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	w := scope.NewWriterBuffer(1024)
	w.Append([]byte("hello"))

	scope.ResetWriterBufferByCapacity(w, 3)

	if w.Len() != 3 {
		t.Fatalf("expected len 3, got %d", w.Len())
	}
	if w.Cap() != 3 {
		t.Fatalf("expected logical cap 3, got %d", w.Cap())
	}
	if string(w.buf[:w.idx]) != "hel" {
		t.Fatalf("expected preserved prefix %q, got %q", "hel", string(w.buf[:w.idx]))
	}
}

func TestScopeResetWriterBufferByCapacityShrinksLogicalWindowWithinBackingArray(t *testing.T) {
	pool := New(DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	w := scope.NewWriterBuffer(512)
	w.Append([]byte("hello"))

	scope.ResetWriterBufferByCapacity(w, 8)

	if w.Len() != 5 {
		t.Fatalf("expected len 5, got %d", w.Len())
	}
	if w.Cap() != 8 {
		t.Fatalf("expected logical cap 8, got %d", w.Cap())
	}
	if string(w.buf[:w.idx]) != "hello" {
		t.Fatalf("expected preserved data %q, got %q", "hello", string(w.buf[:w.idx]))
	}
}

func TestScopeResetWriterBufferByCapacityReplacesBackingArrayWhenNeeded(t *testing.T) {
	pool := New(DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	w := scope.NewWriterBuffer(512)
	w.Append([]byte("hello"))
	oldBuf := w.buf

	scope.ResetWriterBufferByCapacity(w, 2048)

	if w.Len() != 5 {
		t.Fatalf("expected len 5, got %d", w.Len())
	}
	if w.Cap() != 2048 {
		t.Fatalf("expected logical cap 2048, got %d", w.Cap())
	}
	if &w.buf[0] == &oldBuf[0] {
		t.Fatalf("expected backing array to be replaced")
	}
	if string(w.buf[:w.idx]) != "hello" {
		t.Fatalf("expected preserved data %q, got %q", "hello", string(w.buf[:w.idx]))
	}
}

func TestScopeResetWriterBufferByCapacityUsesFullLogicalWindowWhenCapacityEqualsBackingCap(t *testing.T) {
	pool := New(DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	w := scope.NewWriterBuffer(512)
	w.Append([]byte("hello"))
	full := cap(w.buf)

	scope.ResetWriterBufferByCapacity(w, full)

	if w.Len() != 5 {
		t.Fatalf("expected len 5, got %d", w.Len())
	}
	if w.Cap() != full {
		t.Fatalf("expected logical cap %d, got %d", full, w.Cap())
	}
	if string(w.buf[:w.idx]) != "hello" {
		t.Fatalf("expected preserved data %q, got %q", "hello", string(w.buf[:w.idx]))
	}
}

func TestScopeResetWriterBufferByCapacityPanicsOnNonPositiveCapacity(t *testing.T) {
	pool := New(DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	w := scope.NewWriterBuffer(512)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for non-positive capacity")
		}
	}()

	scope.ResetWriterBufferByCapacity(w, 0)
}
