package mempool

import "testing"

func TestGetReturnsSizedSliceWithBucketCapacity(t *testing.T) {
	pool := New(DefaultOptions())
	buf := pool.Get(1500)

	if len(buf) != 1500 {
		t.Fatalf("len(buf) = %d, want 1500", len(buf))
	}
	if cap(buf) != 2048 {
		t.Fatalf("cap(buf) = %d, want 2048", cap(buf))
	}
}

func TestGetOversizeReturnsExactCapacity(t *testing.T) {
	pool := New(DefaultOptions())
	buf := pool.Get(600000)

	if len(buf) != 600000 || cap(buf) != 600000 {
		t.Fatalf("oversize buffer = len %d cap %d, want exact 600000", len(buf), cap(buf))
	}
}

func TestPutReusesBucketedBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	buf := pool.Get(1024)
	buf[0] = 7
	pool.Put(buf)

	got := pool.Get(1000)
	if cap(got) != 1024 {
		t.Fatalf("cap(got) = %d, want 1024", cap(got))
	}
}

func TestPutDropsOversizeBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	buf := make([]byte, 600000)
	pool.Put(buf)

	got := pool.Get(600000)
	if cap(got) != 600000 {
		t.Fatalf("cap(got) = %d, want 600000", cap(got))
	}
}
