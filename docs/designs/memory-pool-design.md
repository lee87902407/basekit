# 新内存池设计方案（实现工作区版）

## 目标

本工作区中的 `mempool` 模块实现基于已确认的设计约束：

1. 底层使用 `sync.Pool` 实现 `[]byte` 分桶池。
2. bucket 最大仅到 `512KB`。
3. 超过 `512KB` 的请求直接申请精确大小的 `[]byte`，归还时直接丢弃。
4. `mempool` 保持底层能力定位：
   - 提供 `BytePool`
   - 提供 `WriterBuffer`/`ReaderBuffer` 双类型模型
   - 提供请求级 `Scope`
5. `WriterBuffer` 写接口放在 `mempool/writer_buffer.go` 中，`ReaderBuffer` 只读接口放在 `mempool/reader_buffer.go` 中。
6. 完整 `ByteBuffer` 风格工具集不放入 `mempool`，未来应独立为更高层的基础封装模块。

## 双类型模型

`mempool` 采用显式双类型模型管理缓冲区生命周期：

### WriterBuffer

- 由 `Scope.NewWriterBuffer(size)` 创建，纳入 Scope 管理。
- 负责写阶段的追加与重置。
- 当前公开方法以 `Len`、`Cap`、`WriteBytes`、`Reset`、`Append`、`AppendByte` 为主。
- 通过 `Scope.ToReaderBuffer(w)` 转移底层 `[]byte` 所有权到 `ReaderBuffer`，转移后 writer 不应继续使用。

### ReaderBuffer

- 由 `Scope.ToReaderBuffer(w)` 产生，纳入 Scope 管理。
- 是固定大小的只读视图。
- 当前公开方法以 `Len`、`Cap`、`ByteAt(i)` 为主，不暴露底层 `[]byte`。

### Scope 统一释放

- `Scope.Close()` 统一释放所有由该 Scope 管理的 `WriterBuffer`、`ReaderBuffer` 和裸 `[]byte`。
- 不再对外暴露公开 `Release()` 方法。
- 生命周期结束后相关对象不应继续使用；当前实现不承诺所有误用路径都统一 panic。

## 当前实现边界

### 已实现

1. `BytePool` 接口与 `BucketedPool` 默认实现。
2. `WriterBuffer` 的写阶段能力：
   - `Len`
   - `Cap`
   - `WriteBytes`
   - `Reset`
   - `Append`
   - `AppendByte`
3. `ReaderBuffer` 的只读能力：
   - `Len`
   - `Cap`
   - `ByteAt`
4. `Scope` 的统一释放能力。
5. Prometheus 兼容指标输出。

### 未纳入当前模块

以下内容明确不在 `mempool` 当前职责内：

1. 完整数值读写、索引、字符串解析等 `ByteBuffer` 风格工具集。
2. 跨异步边界自动生命周期追踪。
3. 自定义 arena / slab 分配器。
4. 零拷贝跨 Go/C 生命周期管理。

## Prometheus 指标

当前模块对外提供 Prometheus 兼容指标，重点包括：

1. `mempool_get_total`：按 bucket 与 pooled 标签统计申请次数。
2. `mempool_requests_total`：总请求量。
3. `mempool_releases_total`：总释放量。
4. `mempool_requests_per_second`：当前秒级请求速率近似值。
5. `mempool_drop_total`：按原因统计丢弃次数。

## 后续演进方向

如果未来需要完整 `ByteBuffer` 风格能力，应新建一个独立模块，定位为：

- 对 `mempool` 的完整上层封装
- 对其他基础模块的组合封装
- 更偏向"基础设施聚合层"，而不是"底层池化层"
