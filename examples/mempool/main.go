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
	buf := pool.Get(1500)
	copy(buf, []byte("hello mempool"))

	fmt.Printf("len=%d cap=%d prefix=%q\n", len(buf), cap(buf), string(buf[:13]))

	pool.Put(buf)

	text, err := stats.GatherText()
	if err != nil {
		panic(err)
	}

	fmt.Println(text)
}
