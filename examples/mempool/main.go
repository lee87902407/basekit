package main

import (
	"fmt"

	"github.com/lee87902407/basekit/mempool"
)

func main() {
	pool := mempool.New(mempool.DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	// WriterBuffer 不自动扩容，容量需要按写入量预估。
	w := scope.NewWriterBuffer(len("hello mempool"))
	w.Append([]byte("hello mempool"))
	fmt.Printf("writer len=%d cap=%d\n", w.Len(), w.Cap())

	// 转移所有权到 ReaderBuffer，后续统一由 scope.Close() 归还。
	r := scope.ToReaderBuffer(w)

	fmt.Printf("reader len=%d cap=%d\n", r.Len(), r.Cap())
}
