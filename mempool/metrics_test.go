package mempool

import (
	"strings"
	"testing"
)

type getEvent struct {
	size   int
	bucket int
	pooled bool
}

type putEvent struct {
	capacity int
	bucket   int
	pooled   bool
}

type dropEvent struct {
	capacity int
	reason   string
}

type recordingStats struct {
	gets  []getEvent
	puts  []putEvent
	drops []dropEvent
}

func (s *recordingStats) OnGet(size int, bucket int, pooled bool) {
	s.gets = append(s.gets, getEvent{size: size, bucket: bucket, pooled: pooled})
}

func (s *recordingStats) OnPut(capacity int, bucket int, pooled bool) {
	s.puts = append(s.puts, putEvent{capacity: capacity, bucket: bucket, pooled: pooled})
}

func (s *recordingStats) OnDrop(capacity int, reason string) {
	s.drops = append(s.drops, dropEvent{capacity: capacity, reason: reason})
}

func TestStatsHooksCalledThroughCurrentScopeAPI(t *testing.T) {
	stats := &recordingStats{}
	opts := DefaultOptions()
	opts.Stats = stats
	pool := New(opts)

	firstScope := pool.NewScope()
	raw := firstScope.Get(1024)
	if len(raw) != 1024 {
		t.Fatalf("unexpected raw length: %d", len(raw))
	}

	writer := firstScope.NewWriterBuffer(2048)
	writer.Append([]byte("ok"))
	firstScope.Close()

	oversizeScope := pool.NewScope()
	oversize := oversizeScope.Get(opts.MaxPooledCap + 1)
	if len(oversize) != opts.MaxPooledCap+1 {
		t.Fatalf("unexpected oversize length: %d", len(oversize))
	}
	oversizeScope.Close()

	secondScope := pool.NewScope()
	reused := secondScope.Get(1024)
	if len(reused) != 1024 {
		t.Fatalf("unexpected reused length: %d", len(reused))
	}
	secondScope.Close()

	if len(stats.gets) != 4 {
		t.Fatalf("unexpected get calls: %+v", stats.gets)
	}
	if len(stats.puts) != 3 {
		t.Fatalf("unexpected put calls: %+v", stats.puts)
	}
	if len(stats.drops) != 1 {
		t.Fatalf("unexpected drop calls: %+v", stats.drops)
	}

	if stats.gets[0] != (getEvent{size: 1024, bucket: 1024, pooled: false}) {
		t.Fatalf("unexpected first raw get event: %+v", stats.gets[0])
	}
	if stats.gets[1] != (getEvent{size: 2048, bucket: 2048, pooled: false}) {
		t.Fatalf("unexpected writer get event: %+v", stats.gets[1])
	}
	if stats.gets[2] != (getEvent{size: opts.MaxPooledCap + 1, bucket: opts.MaxPooledCap + 1, pooled: false}) {
		t.Fatalf("unexpected oversize get event: %+v", stats.gets[2])
	}
	if stats.gets[3] != (getEvent{size: 1024, bucket: 1024, pooled: true}) {
		t.Fatalf("unexpected reused get event: %+v", stats.gets[3])
	}

	if stats.puts[0] != (putEvent{capacity: 2048, bucket: 2048, pooled: true}) {
		t.Fatalf("unexpected first put event: %+v", stats.puts[0])
	}
	if stats.puts[1] != (putEvent{capacity: 1024, bucket: 1024, pooled: true}) {
		t.Fatalf("unexpected raw put event: %+v", stats.puts[1])
	}
	if stats.puts[2] != (putEvent{capacity: 1024, bucket: 1024, pooled: true}) {
		t.Fatalf("unexpected reused put event: %+v", stats.puts[2])
	}

	if stats.drops[0] != (dropEvent{capacity: opts.MaxPooledCap + 1, reason: "oversize_or_zero"}) {
		t.Fatalf("unexpected drop event: %+v", stats.drops[0])
	}
}

func TestPrometheusStatsExposeMetrics(t *testing.T) {
	stats := NewPrometheusStats()
	defer stats.Close()

	opts := DefaultOptions()
	opts.Stats = stats
	pool := New(opts)

	scope := pool.NewScope()
	buf := scope.Get(1500)
	// 请求 1500，bucket 匹配到 2048，返回 len=cap=2048 的数组
	if len(buf) != 2048 {
		t.Fatalf("unexpected pooled length: %d", len(buf))
	}
	oversize := scope.Get(opts.MaxPooledCap + 1)
	// 超出 maxPooledCap 的请求，直接分配请求大小，len=cap=请求大小
	if len(oversize) != opts.MaxPooledCap+1 {
		t.Fatalf("unexpected oversize length: %d", len(oversize))
	}
	writer := scope.NewWriterBuffer(2048)
	writer.AppendByte('a')
	scope.Close()

	text, err := stats.GatherText()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	checks := []string{
		"mempool_get_total",
		"mempool_requests_total",
		"mempool_releases_total",
		"mempool_requests_per_second",
		"mempool_drop_total",
		"bucket=\"2048\"",
		"reason=\"oversize_or_zero\"",
	}

	for i := range checks {
		if !strings.Contains(text, checks[i]) {
			t.Fatalf("metrics output missing %q\n%s", checks[i], text)
		}
	}
}
