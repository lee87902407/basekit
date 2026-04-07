package mempool

import "testing"

func TestBufferImplementsBufferLifecycle(t *testing.T) {
	var _ Buffer = (*HeapBuffer)(nil)
	var _ bufferLifecycle = (*HeapBuffer)(nil)
}

func TestHeapBufferType(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewHeapBuffer(pool, 16)

	if b.Type() != BufferTypeHeap {
		t.Fatalf("buffer type = %d, want %d", b.Type(), BufferTypeHeap)
	}

	b.Release()
}

func TestScopeUseAfterClosePanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	scope.Close()

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "GetHeapBuffer", fn: func() { scope.GetHeapBuffer(1) }},
		{name: "NewBuffer", fn: func() { scope.NewBuffer(1) }},
	}

	for i := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %s", tests[i].name)
				}
			}()
			tests[i].fn()
		}()
	}
}

func TestScopeCloseReleasesTrackedBuffers(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	b1 := scope.NewBuffer(100)
	b2 := scope.NewBuffer(200)

	if len(scope.buffers) != 2 {
		t.Fatalf("tracked buffers = %d, want 2", len(scope.buffers))
	}

	scope.Close()

	if !b1.Released() || !b2.Released() {
		t.Fatal("tracked buffers should be released on Close")
	}
}

func TestScopeCloseReleasesTrackedRawBuffers(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	raw := scope.GetHeapBuffer(1000)
	scope.Close()

	got := pool.Get(900)
	if cap(got) != cap(raw) {
		t.Fatalf("cap(got) = %d, want %d", cap(got), cap(raw))
	}
}
