# mempool 示例文档

## 模块用途

`mempool` 提供一个基于 `sync.Pool` 的 `[]byte` 分桶内存池，用于降低高频短生命周期缓冲区在 Go 中造成的分配与 GC 压力。

## 典型使用场景

1. 高频网络包解析。
2. 协议解码、拼包、序列化。
3. RocksDB 写入前的临时 `[]byte` 编码缓冲。
4. 请求级批量申请与统一释放。

## 轻量 buffer 模型

`mempool` 当前区分两个缓冲区对象：

1. **`WriterBuffer`**：由 `scope.NewWriterBuffer()` 创建，负责写阶段的追加与重置。
2. **`ReaderBuffer`**：由 `scope.ToReaderBuffer(w)` 产生，只提供长度与容量信息，不暴露底层字节切片。

`scope.ToReaderBuffer(w)` 会把底层 `[]byte` 的所有权从 writer 转移给 reader。转换完成后，原 `WriterBuffer` 不应再继续用于写入。

buffer 的底层内存统一由 `Scope.Close()` 归还，不再对外暴露公开 `Release()`。

## 最小可运行示例

```go
package main

import (
	"fmt"

	"github.com/lee87902407/basekit/mempool"
)

func main() {
	pool := mempool.New(mempool.DefaultOptions())
	scope := pool.NewScope()
	defer scope.Close()

	// WriterBuffer 不会自动扩容，容量需要按写入量预估。
	w := scope.NewWriterBuffer(len("hello mempool"))
	w.Append([]byte("hello mempool"))
	fmt.Printf("writer len=%d cap=%d\n", w.Len(), w.Cap())

	// 转移所有权到 ReaderBuffer，后续统一由 scope.Close() 归还。
	r := scope.ToReaderBuffer(w)

	fmt.Printf("reader len=%d cap=%d\n", r.Len(), r.Cap())
}
```

## WriterBuffer 写接口

当前 `WriterBuffer` 提供以下公开方法：

1. `Bytes() []byte`：返回当前底层字节切片。
1. `Len() int`：返回当前长度。
2. `Cap() int`：返回创建时声明的容量。
3. `Reset()`：保留底层数组，仅重置长度与写入位置。
4. `Append([]byte)`：追加一段字节；若超出剩余容量会直接 panic。
5. `AppendByte(byte)`：追加单个字节；若超出剩余容量会直接 panic。
6. `CloneByBuffer() *WriterBuffer`：在同一 `Scope` 下创建当前内容的独立副本。

`WriterBuffer` 当前**不会自动扩容**。如果写入数据量超过创建时申请的容量，`Append` / `AppendByte` 会直接触发 `panic("mempool: buffer overflow")`。

## ReaderBuffer 只读接口

`ReaderBuffer` 提供以下只读方法：

1. `Len() int`：返回缓冲区长度。
2. `Cap() int`：返回底层切片容量。

当前 `ReaderBuffer` 不提供 `Bytes()`、`Clone()`、`DetachedCopy()` 等接口。如需读取具体内容，应在转为 `ReaderBuffer` 之前自行拷贝或消费 `WriterBuffer.Bytes()` 返回的数据。

## 接入注意事项

1. bucket 最大只到 `512KB`。
2. 超过 `512KB` 的请求会直接分配精确大小的 `[]byte`，归还时直接丢弃，不进入池。
3. `Put` 按 `cap(buf)` 判断归属 bucket，而不是按 `len(buf)`。
4. buffer 在跨异步边界或跨 cgo 生命周期时，必须确认 ownership 后才能归还。
5. `Scope` 通过 `pool.NewScope()` 创建，不提供 `mempool.NewScope(...)` 这类包级构造函数。
6. `scope.ToReaderBuffer(w)` 之后，不应再把 `w` 当作可写缓冲区继续使用。
7. `Append` / `AppendByte` 不会触发自动扩容；容量不足会直接 panic，因此创建 `WriterBuffer` 时要预估好容量。
8. `WriterBuffer.Len()` / `Cap()` 与 `ReaderBuffer.Len()` / `Cap()` 可视为读操作；对象在所有权转移或 `Scope.Close()` 后，这些读取结果可能表现为 `nil`、`0` 或静默返回，文档不应假定“所有访问都会 panic”。

## 失效对象行为

以下行为应视为误用：

1. `WriterBuffer` 在 `scope.ToReaderBuffer(w)` 之后继续写入。
2. `ReaderBuffer` 在 `Scope.Close()` 之后继续依赖其容量或长度信息。
3. `WriterBuffer` 在 `Scope.Close()` 之后继续写入。

当前实现下，写操作误用更容易暴露为 panic；读操作则可能只返回 `nil`、`0` 或静默成功。调用方应在生命周期边界前完成消费，不要依赖失效后的具体表现。

## 与 README 的跳转关系

1. 统一入口位于 [`README.md`](../../README.md)。
2. 中文说明位于 [`README.zh-CN.md`](../../README.zh-CN.md)。
3. 英文说明位于 [`README.en.md`](../../README.en.md)。
4. 对应示例代码位于 [`examples/mempool/`](../../examples/mempool/)。
