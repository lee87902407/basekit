# mempool 示例文档

## 模块用途

`mempool` 提供一个基于 `sync.Pool` 的 `[]byte` 分桶内存池，用于降低高频短生命周期缓冲区在 Go 中造成的分配与 GC 压力。

## 典型使用场景

1. 高频网络包解析。
2. 协议解码、拼包、序列化。
3. RocksDB 写入前的临时 `[]byte` 编码缓冲。
4. 请求级批量申请与统一释放。

## 轻量 buffer 模型

`mempool` 区分两个缓冲区对象：

1. **`WriterBuffer`**：由 `Scope.NewWriterBuffer()` 创建，负责写阶段的追加、扩容、重置与构建。
2. **`ReaderBuffer`**：由 `WriterBuffer.ToReaderBuffer()` 产生，提供弱只读访问；`Bytes()` 直接返回底层 `[]byte`，调用者不得修改返回的切片内容。

`ToReaderBuffer()` 会把底层 `[]byte` 的所有权从 writer 转移给 reader。转换完成后，原 `WriterBuffer` 立即失效，访问会触发 panic。

buffer 的底层内存统一由 `Scope.Close()` 归还，不再对外暴露公开 `Release()`。`Scope.Close()` 后所有相关缓冲区均失效，访问会触发 panic。

## 最小可运行示例

```go
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
```

## WriterBuffer 写接口

当前 `WriterBuffer` 提供以下写阶段方法：

1. `Reset()`：保留容量，仅重置长度。
2. `Resize(n int)`：将缓冲区长度调整为 n。
3. `EnsureCapacity(additional int)`：在追加前确保底层容量足够，不够时通过池重新获取并迁移数据。
4. `Append([]byte)` / `AppendByte(byte)`：先做受控扩容，再执行写入。
5. `Clone() []byte`：返回当前内容的独立副本。
6. `DetachedCopy() []byte`：语义与 Clone 一致。
7. `ToReaderBuffer() *ReaderBuffer`：转移所有权到只读缓冲区。

## ReaderBuffer 只读接口

`ReaderBuffer` 提供以下只读方法：

1. `Bytes() []byte`：返回底层字节切片（弱只读，直接返回底层 `[]byte`）。
2. `Len() int`：返回缓冲区长度。
3. `Cap() int`：返回缓冲区容量。
4. `Released() bool`：返回缓冲区是否已释放。
5. `Clone() []byte`：返回当前内容的独立副本。
6. `DetachedCopy() []byte`：语义与 Clone 一致。

## 接入注意事项

1. bucket 最大只到 `512KB`。
2. 超过 `512KB` 的请求会直接分配精确大小的 `[]byte`，归还时直接丢弃，不进入池。
3. `Put` 按 `cap(buf)` 判断归属 bucket，而不是按 `len(buf)`。
4. buffer 在跨异步边界或跨 cgo 生命周期时，必须确认 ownership 后才能归还。
5. `WriterBuffer.ToReaderBuffer()` 之后原 writer 立即失效，不能再调用任何方法。
6. `ReaderBuffer.Bytes()` 是弱只读，直接返回底层 `[]byte`，调用者不应修改返回的切片内容。
7. `Append` / `AppendByte` 不依赖 Go 内置 `append` 的隐式扩容，而是优先走池化扩容策略。
8. debug 与非 debug 构建下，已释放的缓冲区对象均不可继续使用，访问会触发 panic。

## 失效对象行为

无论是 debug 还是非 debug 构建，以下行为均会触发 panic：

1. `WriterBuffer` 在 `ToReaderBuffer()` 之后继续调用任何方法。
2. `ReaderBuffer` 在 `Scope.Close()` 之后继续调用任何方法。
3. `WriterBuffer` 在 `Scope.Close()` 之后继续调用任何方法。

即：已释放或已失效的缓冲区对象在任何构建模式下都不可继续使用。

## debug 构建检查

在排查 buffer 误用时，可以显式开启 `debug` 构建标签：

```bash
go test -tags debug ./mempool
go test -tags debug ./...
```

`debug` 标签的价值在于提供更早、更丰富的误用暴露，而非让失效对象 panic 成为 debug 专属行为。

## 与 README 的跳转关系

1. 统一入口位于 [`README.md`](../../README.md)。
2. 中文说明位于 [`README.zh-CN.md`](../../README.zh-CN.md)。
3. 英文说明位于 [`README.en.md`](../../README.en.md)。
4. 对应示例代码位于 [`examples/mempool/`](../../examples/mempool/)。
