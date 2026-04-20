package mempool

import "testing"

func TestScopeCloseReleasesTrackedWriterBuffers(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w1 := scope.NewWriterBuffer(100)
	w2 := scope.NewWriterBuffer(200)

	scope.Close()

	if !w1.Released() || !w2.Released() {
		t.Fatal("tracked writer buffers should be released on Close")
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

func TestScopeNewWriterBufferAfterClosePanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on NewWriterBuffer after Close")
		}
	}()
	scope.NewWriterBuffer(32)
}

func TestScopeGetAfterClosePanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on Get after Close")
		}
	}()
	scope.Get(32)
}

func TestScopeTrackAfterClosePanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on Track after Close")
		}
	}()
	scope.Track(make([]byte, 10))
}
