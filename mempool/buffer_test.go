//go:build debug

package mempool

import (
	"bytes"
	"testing"
)

func TestBufferReleaseReturnsToPoolOnce(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 1000)

	b.Append([]byte("abc"))
	b.Release()

	if !b.Released() {
		t.Fatal("buffer should be marked released")
	}
}

func TestBufferReleaseTwicePanics(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 32)
	b.Release()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on double release")
		}
	}()

	b.Release()
}

func TestBufferUseAfterReleasePanics(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 32)
	b.Release()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on use after release")
		}
	}()

	b.AppendByte('x')
}

func TestBufferAppendGrowsThroughPoolWhenCapacityInsufficient(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 510)
	b.Reset()
	b.Append(bytes.Repeat([]byte{'a'}, 510))

	oldCap := b.Cap()
	b.Append(bytes.Repeat([]byte{'b'}, 10))

	if b.Cap() <= oldCap {
		t.Fatalf("buffer cap did not grow, old=%d new=%d", oldCap, b.Cap())
	}
	if b.Len() != 520 {
		t.Fatalf("buffer len = %d, want 520", b.Len())
	}
	if b.Cap() != 1024 {
		t.Fatalf("buffer cap = %d, want pooled cap 1024", b.Cap())
	}
	for i := 0; i < 510; i++ {
		if b.Bytes()[i] != 'a' {
			t.Fatalf("buffer content mismatch at %d", i)
		}
	}
	for i := 510; i < 520; i++ {
		if b.Bytes()[i] != 'b' {
			t.Fatalf("buffer content mismatch at %d", i)
		}
	}
}

func TestBufferAppendByteGrowsThroughPoolWhenCapacityInsufficient(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 1)
	b.Reset()
	for i := 0; i < 600; i++ {
		b.AppendByte('a')
	}

	if b.Len() != 600 {
		t.Fatalf("buffer len = %d, want 600", b.Len())
	}
	if b.Cap() < 600 {
		t.Fatalf("buffer cap = %d, want >= 600", b.Cap())
	}
	if b.Cap() != 1024 {
		t.Fatalf("buffer cap = %d, want pooled cap 1024", b.Cap())
	}
}

func TestBufferCloneCreatesDetachedCopy(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 10)
	b.Reset()
	b.Append([]byte("hello"))

	dup := b.Clone()
	b.Append([]byte("-world"))

	if string(dup) != "hello" {
		t.Fatalf("clone mismatch: %q", string(dup))
	}
}

func TestBufferEnsureCapacity(t *testing.T) {
	pool := New(DefaultOptions())
	b := NewBuffer(pool, 16)
	b.Reset()
	b.Append([]byte("abc"))
	b.EnsureCapacity(900)

	if b.Cap() != 1024 {
		t.Fatalf("buffer cap = %d, want 1024", b.Cap())
	}
	if string(b.Bytes()) != "abc" {
		t.Fatalf("buffer content mismatch after ensure: %q", string(b.Bytes()))
	}
}
