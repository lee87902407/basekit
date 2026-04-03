package mempool

import (
	"os"
	"strings"
	"testing"
)

func TestBucketForRoundsUpToNearestClass(t *testing.T) {
	pool := New(DefaultOptions())

	tests := []struct {
		name string
		size int
		want int
	}{
		{name: "1 byte", size: 1, want: 512},
		{name: "513 bytes", size: 513, want: 1024},
		{name: "4k exact", size: 4096, want: 4096},
		{name: "64k plus one", size: 65537, want: 131072},
		{name: "512k exact", size: 524288, want: 524288},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pool.Bucket(tt.size); got != tt.want {
				t.Fatalf("Bucket(%d) = %d, want %d", tt.size, got, tt.want)
			}
		})
	}
}

func TestBucketForOversizeReturnsOriginalSize(t *testing.T) {
	pool := New(DefaultOptions())

	if got := pool.Bucket(524289); got != 524289 {
		t.Fatalf("Bucket(524289) = %d, want 524289", got)
	}
}

func TestBucketImplementationAvoidsRangeValueBinding(t *testing.T) {
	data, err := os.ReadFile("bucket.go")
	if err != nil {
		t.Fatalf("read bucket.go: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "for _, class := range p.classes") {
		t.Fatal("bucket.go should not use range value binding over p.classes")
	}
}
