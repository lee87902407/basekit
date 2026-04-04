package mempool

import "testing"

func TestBufferImplementsBufferLifecycle(t *testing.T) {
	var _ bufferLifecycle = (*Buffer)(nil)
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
	raw := scope.Get(1000)
	scope.Close()

	got := pool.Get(900)
	if cap(got) != cap(raw) {
		t.Fatalf("cap(got) = %d, want %d", cap(got), cap(raw))
	}
}
