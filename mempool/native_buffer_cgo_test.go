package mempool

import (
	"testing"
	"unsafe"
)

type nativeBufferTestPool struct {
	puts int
}

func (p *nativeBufferTestPool) Get(size int) []byte {
	return make([]byte, size)
}

func (p *nativeBufferTestPool) Put(buf []byte) {
	p.puts++
}

func (p *nativeBufferTestPool) Bucket(size int) int {
	return size
}

func (p *nativeBufferTestPool) MaxPooledCap() int {
	return 0
}

func TestNativeBufferImplementsBufferLifecycle(t *testing.T) {
	var _ Buffer = (*NativeBuffer)(nil)
	var _ bufferLifecycle = (*NativeBuffer)(nil)
}

func TestNativeBufferCloneAndDetachedCopy(t *testing.T) {
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)

	clone := b.Clone()
	detached := b.DetachedCopy()
	b.Release()

	if string(clone) != "native" {
		t.Fatalf("clone = %q, want native", string(clone))
	}
	if string(detached) != "native" {
		t.Fatalf("detached = %q, want native", string(detached))
	}
	if !b.Released() {
		t.Fatal("native buffer should be marked released")
	}
	if b.Bytes() != nil {
		t.Fatal("native buffer bytes should detach on release")
	}
}

func TestNativeBufferType(t *testing.T) {
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)

	if b.Type() != BufferTypeNative {
		t.Fatalf("buffer type = %d, want %d", b.Type(), BufferTypeNative)
	}

	b.Release()
}

func TestNativeBufferWriteOperationsPanic(t *testing.T) {
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "Reset", fn: func() { b.Reset() }},
		{name: "EnsureCapacity", fn: func() { b.EnsureCapacity(1) }},
		{name: "Resize", fn: func() { b.Resize(1) }},
		{name: "Append", fn: func() { b.Append([]byte("x")) }},
		{name: "AppendByte", fn: func() { b.AppendByte('x') }},
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

	b.Release()
}

func TestNativeBufferReleaseCallsFreeFunc(t *testing.T) {
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	freed := 0
	b := NewNativeBuffer(owner, data, size, func(ptr unsafe.Pointer) {
		freed++
		freeOwnedNativeBufferForTesting(ptr)
	})

	b.Release()

	if freed != 1 {
		t.Fatalf("free function call count = %d, want 1", freed)
	}
}

func TestScopeCloseReleasesNativeBufferWithoutPoolPut(t *testing.T) {
	pool := &nativeBufferTestPool{}
	scope := NewScope(pool)
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := scope.NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)

	scope.Close()

	if !b.Released() {
		t.Fatal("native buffer should be released on scope close")
	}
	if pool.puts != 0 {
		t.Fatalf("pool put count = %d, want 0", pool.puts)
	}
}

func TestScopeGetNativeBufferTracksBuffer(t *testing.T) {
	pool := &nativeBufferTestPool{}
	scope := NewScope(pool)
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))
	b := scope.GetNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)

	if len(scope.buffers) != 1 {
		t.Fatalf("tracked buffers = %d, want 1", len(scope.buffers))
	}
	if b.Type() != BufferTypeNative {
		t.Fatalf("buffer type = %d, want %d", b.Type(), BufferTypeNative)
	}

	scope.Close()
}

func TestScopeNewNativeBufferAfterClosePanics(t *testing.T) {
	pool := &nativeBufferTestPool{}
	scope := NewScope(pool)
	scope.Close()
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on native buffer creation after close")
		}
		freeOwnedNativeBufferForTesting(owner)
	}()

	scope.NewNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)
}

func TestScopeGetNativeBufferAfterClosePanics(t *testing.T) {
	pool := &nativeBufferTestPool{}
	scope := NewScope(pool)
	scope.Close()
	owner, data, size := newOwnedNativeBufferForTesting([]byte("native"))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on native buffer get after close")
		}
		freeOwnedNativeBufferForTesting(owner)
	}()

	scope.GetNativeBuffer(owner, data, size, nativeBufferTestFreeFunc)
}
