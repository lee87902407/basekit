//go:build !debug

package mempool

import "testing"

func TestBufferReleaseTwiceIgnoredWithoutDebug(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewHeapBuffer(pool, 32)
	b.Release()

	b.Release()
}

func TestBufferUseAfterReleaseIgnoredWithoutDebug(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewHeapBuffer(pool, 32)
	b.Release()

	b.AppendByte('x')

	if b.Len() != 1 {
		t.Fatalf("buffer len = %d, want 1", b.Len())
	}
	if b.Released() {
		t.Fatal("buffer should become reusable again after non-debug use")
	}
}

func TestScopeCloseStillReleasesBufferReusedAfterReleaseWithoutDebug(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	b := scope.NewBuffer(32)
	b.Release()
	b.AppendByte('x')
	wantCap := b.Cap()

	scope.Close()

	if !b.Released() {
		t.Fatal("buffer should be marked released after scope close")
	}

	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}
}
