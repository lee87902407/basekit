package main

import (
	"fmt"

	"github.com/lee87902407/basekit/mempool"
)

func main() {
	stats := mempool.NewPrometheusStats()
	defer stats.Close()

	opts := mempool.DefaultOptions()
	opts.Stats = stats
	pool := mempool.New(opts)
	heap := mempool.NewHeapBuffer(pool, 32)
	defer heap.Release()
	heap.Reset()
	heap.Append([]byte("hello mempool"))

	fmt.Printf("heap-len=%d heap-cap=%d text=%q\n", heap.Len(), heap.Cap(), string(heap.Bytes()))

	raw := pool.Get(1500)
	copy(raw, []byte("raw-bytes"))
	fmt.Printf("raw-len=%d raw-cap=%d\n", len(raw), cap(raw))
	pool.Put(raw)

	text, err := stats.GatherText()
	if err != nil {
		panic(err)
	}

	fmt.Println(text)
}
