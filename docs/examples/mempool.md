# mempool 示例文档

## 模块用途

`mempool` 提供一个基于 `sync.Pool` 的 `[]byte` 分桶内存池，用于降低高频短生命周期缓冲区在 Go 中造成的分配与 GC 压力。

## 典型使用场景

1. 高频网络包解析。
2. 协议解码、拼包、序列化。
3. RocksDB 写入前的临时 `[]byte` 编码缓冲。
4. 请求级批量申请与统一释放。

## 最小可运行示例

```go
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

	fmt.Printf("len=%d cap=%d text=%q\n", heap.Len(), heap.Cap(), string(heap.Bytes()))

	raw := pool.Get(1500)
	copy(raw, []byte("raw-bytes"))
	fmt.Printf("raw-len=%d raw-cap=%d\n", len(raw), cap(raw))
	pool.Put(raw)

	text, _ := stats.GatherText()
	fmt.Println(text)
}
```

## 轻量工具方法

`Buffer` 接口当前有两种实现：

1. `HeapBuffer`：使用 `mempool.NewHeapBuffer(pool, size)` 创建，可写、可扩容、释放后归还池。
2. `NativeBuffer`：在 cgo 场景下使用 `mempool.NewNativeBuffer(owner, data, size, freeFun)`、`Scope.NewNativeBuffer(...)` 或 `Scope.GetNativeBuffer(...)` 创建，只读、释放时调用注入的 `freeFun`，不会回收到 Go 的 byte pool。

可以通过 `Buffer.Type()` 区分具体来源：

1. `BufferTypeHeap`
2. `BufferTypeNative`

其中 `HeapBuffer` 提供的轻量工具方法包括：

1. `EnsureCapacity(additional int)`：在追加前确保底层容量足够，不够时通过池重新获取并迁移数据。
2. `Clone()`：返回当前内容的独立副本。
3. `Reset()`：保留容量，仅重置长度。
4. `Append([]byte)` / `AppendByte(byte)`：先做受控扩容，再执行写入。

## 接入注意事项

1. bucket 最大只到 `512KB`。
2. 超过 `512KB` 的请求会直接分配精确大小的 `[]byte`，归还时直接丢弃，不进入池。
3. `Put` 按 `cap(buf)` 判断归属 bucket，而不是按 `len(buf)`。
4. buffer 在跨异步边界或跨 cgo 生命周期时，必须确认 ownership 后才能归还。
5. `HeapBuffer.Release()` 之后不能继续持有其底层数据引用；`NativeBuffer.Release()` 会调用注入的 `freeFun` 并断开对原生内存的 Go 视图。
6. `Append` / `AppendByte` 不依赖 Go 内置 `append` 的隐式扩容，而是优先走池化扩容策略。
7. `NativeBuffer` 是只读包装，`Reset`、`Resize`、`EnsureCapacity`、`Append`、`AppendByte` 等写入类方法会直接 panic。
8. 默认构建下，`HeapBuffer` 不会因为 `use after release` 或重复 `Release` 直接 panic；如果释放后又继续写入，它会恢复为可继续托管、可被后续 `Scope.Close()` 回收的状态。如需在开发阶段开启更严格检查，请使用 `-tags debug`。
9. `Scope.GetHeapBuffer` 用于申请并托管原始 heap `[]byte`；`Track` 已移除。
10. `Scope.Close()` 之后不能继续调用 `GetHeapBuffer`、`NewBuffer`、`NewNativeBuffer` 或 `GetNativeBuffer`；这类操作会直接 panic，避免新增资源脱离请求级释放语义。

## debug 构建检查

在排查 buffer 误用时，可以显式开启 `debug` 构建标签：

```bash
go test -tags debug ./mempool
go test -tags debug ./...
```

开启后，`Buffer` 生命周期会在以下场景直接 panic：

1. `HeapBuffer.Release()` 之后继续调用 `Reset`、`Resize`、`Append`、`AppendByte`、`EnsureCapacity`、`Clone` 等方法。
2. 对同一个 `HeapBuffer` 或 `NativeBuffer` 重复调用 `Release()`。

## 与 README 的跳转关系

1. 统一入口位于 [`README.md`](../../README.md)。
2. 中文说明位于 [`README.zh-CN.md`](../../README.zh-CN.md)。
3. 英文说明位于 [`README.en.md`](../../README.en.md)。
4. 对应示例代码位于 [`examples/mempool/`](../../examples/mempool/)。
