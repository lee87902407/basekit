package main

import (
	"fmt"

	"github.com/lee87902407/basekit/mempool"
)

func main() {
	pool := mempool.New(mempool.DefaultOptions())
	scope := mempool.NewScope(pool)
	defer scope.Close()

	// 创建 WriterBuffer 并写入数据
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("hello mempool"))

	// 转移所有权到 ReaderBuffer
	r := w.ToReaderBuffer()

	fmt.Printf("len=%d cap=%d text=%q\n", r.Len(), r.Cap(), string(r.Bytes()))
}
