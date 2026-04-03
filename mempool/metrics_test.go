package mempool

import (
	"strings"
	"testing"
)

type testStats struct {
	gets  int
	puts  int
	drops int
}

func (s *testStats) OnGet(size int, bucket int, pooled bool)     { s.gets++ }
func (s *testStats) OnPut(capacity int, bucket int, pooled bool) { s.puts++ }
func (s *testStats) OnDrop(capacity int, reason string)          { s.drops++ }

func TestStatsHooksCalled(t *testing.T) {
	stats := &testStats{}
	opts := DefaultOptions()
	opts.Stats = stats
	pool := New(opts)

	buf := pool.Get(1024)
	pool.Put(buf)
	pool.Put(make([]byte, 600000))

	if stats.gets == 0 || stats.puts == 0 || stats.drops == 0 {
		t.Fatalf("stats not called: %+v", *stats)
	}
}

func TestPrometheusStatsExposeMetrics(t *testing.T) {
	stats := NewPrometheusStats()
	defer stats.Close()
	opts := DefaultOptions()
	opts.Stats = stats
	pool := New(opts)

	buf := pool.Get(1500)
	pool.Put(buf)
	pool.Get(600000)

	text, err := stats.GatherText()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	checks := []string{
		"mempool_get_total",
		"mempool_requests_total",
		"mempool_releases_total",
		"mempool_requests_per_second",
		"bucket=\"2048\"",
	}

	for _, item := range checks {
		if !strings.Contains(text, item) {
			t.Fatalf("metrics output missing %q\n%s", item, text)
		}
	}
}
