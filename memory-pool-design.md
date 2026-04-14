# 新内存池设计方案（基于 `sync.Pool` 的分层设计）

## 1. 文档目的

本文档用于在 `basekit` 中设计一个新的 Go 内存池模块，目标场景是：

- 高频网络包解析
- 高频短生命周期 `[]byte` 分配与释放
- RocksDB / cgo 边界前后的临时缓冲复用
- 编码、解码、拼包、批量写入等热点路径

本设计**只定义方案与类设计，不包含代码实现**。

## 2. 输入材料与结论来源

本设计综合了以下两类材料：

1. 仓库内的 `pool设计.md`
   - 明确指出 Go 堆上的 `[]byte` 不受 jemalloc 优化影响。
   - 强调对于 10w/s 级别短命 `[]byte`，应优先考虑“分桶 + sync.Pool”。
   - 重点强调 ownership、超大对象污染、防止异步误归还等边界问题。

2. 外部参考实现 `/Users/yanjie/data/code/work/storage/flight-storage-tools/memory/*.go`
   - 已有一版基于 `sync.Pool` 的分桶池实现。
   - 通过 `ByteBuffer` 封装了 `[]byte`、读写索引和追加/读取能力。
   - 通过 `Chunk` 提供批量申请和统一释放能力。

综合判断：

- 旧实现证明了“`sync.Pool` + 分桶 + 请求级统一回收”是可落地的。
- 但旧实现把“池化”和“编解码工具”耦合到 `ByteBuffer`，边界过重，不利于做成多项目共享基础模块。
- 新设计应保留其**分桶复用**和**批量释放**思路，但将“底层复用能力”和“上层读写包装”拆开。

## 3. 设计目标

### 3.1 功能目标

新内存池需要满足：

1. 支持高频 `[]byte` 复用。
2. 按固定 bucket 进行容量分桶。
3. 提供极简、低开销的底层 `[]byte` API。
4. 提供可选的上层 `Buffer` 包装，承载更安全的使用体验。
5. 提供可选的 `Scope` 批量释放能力，适合单请求/单任务生命周期。
6. 对大对象保持明确边界，避免污染池。
7. 对 RocksDB / cgo / 异步边界给出明确 ownership 规则。

### 3.2 性能目标

1. 降低热点路径 `make([]byte, n)` 次数。
2. 降低 `alloc_objects`、`alloc_space`。
3. 降低 GC 压力与 P99/P999 抖动。
4. 保持多 goroutine 下较低竞争。
5. 优先优化小中型、高频、短命 buffer。

### 3.3 非目标

以下能力**不属于 v1 目标**：

1. 动态学习 bucket 分布并自动重分桶。
2. 自定义 arena/slab 分配器。
3. 跨异步边界的自动生命周期追踪。
4. 零拷贝穿越 Go/C 生命周期管理。
5. 通用对象池（v1 只聚焦 `[]byte`）。

## 4. 方案对比

### 4.1 方案 A：仅提供原生 `[]byte` 池

#### 描述

底层直接暴露：

- `Get(size int) []byte`
- `Put(buf []byte)`

不提供任何包装对象。

#### 优点

1. 性能路径最短。
2. 心智模型简单。
3. 对热点路径最友好。

#### 缺点

1. 容易误用，尤其是异步场景和 ownership 不清晰时。
2. 无法承载请求级批量释放能力。
3. 无法自然挂接调试标记、状态位、来源信息。

#### 适用场景

- 极致性能路径。
- 业务代码非常自律，严格遵守归还规则。

---

### 4.2 方案 B：池直接返回 `ByteBuffer` 风格对象

#### 描述

底层池直接管理带状态的对象，例如：

- `ByteBuffer`
- 包含 `[]byte`、读写索引、append/read 方法、Reset 等

#### 优点

1. 易于封装协议编解码能力。
2. 易于挂接标记、状态与辅助方法。
3. 对使用方更“顺手”。

#### 缺点

1. 池化边界和业务语义耦合过重。
2. 后续容易演变成“万能 buffer 工具类”。
3. 多项目共享时，不同项目对 `Buffer` 语义要求不一致，容易失控。

#### 适用场景

- 单一项目内部。
- 编解码模式高度统一。

---

### 4.3 方案 C：分层设计（推荐）

#### 描述

采用三层结构：

1. **底层 `BytePool`**：只负责 `[]byte` 分桶复用。
2. **上层 `Buffer` 包装**：可选，为需要更强约束的路径提供对象语义。
3. **请求级 `Scope`**：可选，用于统一释放一组从池中借出的对象。

#### 优点

1. 热点路径仍可直接使用原生 `[]byte`。
2. 需要安全性时，可选择更高层抽象。
3. 复用策略和协议读写逻辑解耦。
4. 更适合作为 `basekit` 中的长期基础模块。

#### 缺点

1. 设计比 A 稍复杂。
2. 需要明确三层边界，避免重复能力。

#### 适用场景

- 多项目共享基础库。
- 同时兼顾高性能路径和安全性路径。

### 4.4 最终推荐

**推荐采用方案 C。**

推荐原因：

1. 它保留了 `sync.Pool + bucket` 的性能优势。
2. 它吸收了旧实现中的 `Chunk` 批量释放思路，但不再把读写能力和池对象绑死。
3. 它允许不同项目按需使用：
   - 追求极致性能的路径使用原生 `[]byte`
   - 追求易用和安全的路径使用 `Buffer`
   - 需要请求级统一清理的路径使用 `Scope`

## 5. 总体架构

### 5.1 模块定位

建议未来模块目录使用：

- `mempool/`

原因：

1. 语义更明确，直接表达“内存池”而不是泛化的 pool。
2. 与当前设计目标一致，便于后续对外暴露稳定模块名。
3. 后续可在模块内继续细分 `byte pool`、`scope pool`、`buffer wrapper`。

### 5.2 架构分层

#### L1：底层池化层

负责：

- bucket 配置
- 容量向上取整
- `sync.Pool` 复用
- 大对象丢弃策略
- 基础统计钩子

不负责：

- 协议编解码
- 业务字段读写
- 自动 ownership 推断

#### L2：对象包装层

负责：

- 用对象形式包装 `[]byte`
- 记录所属池与释放状态
- 提供少量“安全辅助能力”

不负责：

- 做成一个超级编解码类
- 承载完整协议序列化框架

#### L3：生命周期聚合层

负责：

- 在单请求/单任务内批量借出对象
- 在结束时统一释放
- 降低多次 `defer Put(...)` 的碎片化管理成本

## 6. 核心类设计

以下为**类/类型设计**，不是代码实现。

### 6.1 `BytePool` 接口

#### 职责

定义底层 `[]byte` 池抽象。

#### 建议接口

```go
type BytePool interface {
    Get(size int) []byte
    Put(buf []byte)
    Bucket(size int) int
    MaxPooledCap() int
}
```

#### 说明

1. `Get(size)` 返回长度为 `size`、容量为最近 bucket 的切片。
2. `Put(buf)` 只接受标准 bucket 容量对象。
3. `Bucket(size)` 主要用于调试、观测和上层包装对象记录来源。
4. `MaxPooledCap()` 用于显式暴露池化边界。

### 6.2 `BucketedPool` 类

#### 职责

`BytePool` 的默认实现。

#### 核心字段

```go
type BucketedPool struct {
    classes       []SizeClass
    maxPooledCap  int
    zeroOnPut     bool
    zeroOnGet     bool
    stats         StatsCollector
}
```

#### 字段说明

- `classes`：容量分桶定义。
- `maxPooledCap`：最大池化容量，超过则直接分配，不回池。
- `zeroOnPut`：归还时是否清零，默认关闭，调试/安全模式可开启。
- `zeroOnGet`：借出时是否清零，默认关闭。
- `stats`：可选统计收集器。

#### 设计说明

1. `BucketedPool` 是默认的生产实现。
2. v1 使用静态 bucket，不做动态学习。
3. 必须保证 bucket 严格有序且无重复。

### 6.3 `SizeClass` 类

#### 职责

表示一个 bucket。

#### 建议结构

```go
type SizeClass struct {
    Cap  int
    Pool sync.Pool
}
```

#### 说明

1. 每个 bucket 对应一个 `sync.Pool`。
2. 池内统一存放 `cap == Cap` 的 `[]byte`。
3. v1 不引入额外锁结构，直接利用 `sync.Pool` 的并发语义。

### 6.4 `Buffer` 类（可选包装对象）

#### 职责

为上层提供更安全的对象语义，但不承担重型编解码职责。

#### 建议结构

```go
type Buffer struct {
    buf       []byte
    pool      BytePool
    pooledCap int
    released  bool
}
```

#### 建议方法分类

1. **基础数据访问**
   - `Bytes() []byte`
   - `Len() int`
   - `Cap() int`

2. **长度/内容管理**
   - `Reset()`
   - `Resize(n int)`
   - `Append(p []byte)`
   - `AppendByte(v byte)`

3. **生命周期控制**
   - `Release()`
   - `DetachedCopy()`

#### 设计边界

`Buffer` **不应该**承载以下能力：

1. 大量协议字段级 `ReadUInt32/AppendInt64/...` 方法。
2. 读写索引状态机。
3. 通用序列化框架。

原因：

- 这些能力会让 `Buffer` 重新变成旧版 `ByteBuffer` 那种“万能对象”。
- 新设计要把“复用”和“编解码”拆开。

#### 允许保留的最小方法集

`Buffer` 可以保留极少量基础便捷方法，前提是这些方法不改变模块边界：

- `Reset`
- `Resize`
- `Append`
- `AppendByte`
- `Bytes`
- `Release`

### 6.5 `Scope` 类（请求级统一释放）

#### 职责

吸收旧实现中 `Chunk` 的优点，为单请求/单任务提供统一释放机制。

#### 建议结构

```go
type Scope struct {
    pool    BytePool
    buffers []*Buffer
    raws    [][]byte
    closed  bool
}
```

#### 建议接口

```go
type Scope interface {
    Get(size int) []byte
    NewBuffer(size int) *Buffer
    Track(buf []byte)
    Close()
}
```

#### 说明

1. `Scope` 适用于单请求或单任务上下文。
2. `Close()` 时统一归还所有被追踪对象。
3. `Track(buf)` 用于把外部获得的原生 `[]byte` 纳入统一清理。
4. `Scope` 本身不是并发安全容器，默认按单协程/单任务使用。

### 6.6 `StatsCollector` 接口（可选）

#### 职责

为观测与压测提供扩展点。

#### 建议接口

```go
type StatsCollector interface {
    OnGet(size int, bucket int, pooled bool)
    OnPut(capacity int, bucket int, pooled bool)
    OnDrop(capacity int, reason string)
}
```

#### 说明

1. 默认实现可为空。
2. 用于接 Prometheus、runtime/metrics 或本地 benchmark 观测。
3. 不应影响主路径性能，必须允许关闭。

## 7. 容量分桶设计

### 7.1 默认 bucket

建议默认 bucket：

- 512B
- 1KB
- 2KB
- 4KB
- 8KB
- 16KB
- 32KB
- 64KB
- 128KB
- 256KB
- 512KB

### 7.2 选择依据

1. 覆盖常见网络包、KV 编码、拼包、协议解析场景。
2. 与 `pool设计.md` 中的建议一致。
3. 粒度足够细，但不会把 bucket 切得过密。

### 7.3 最大池化阈值

默认建议：

- `512KB`

设计约束：

- bucket 最大只到 `512KB`
- 当请求大小超过 `512KB` 时，直接申请一个精确大小的 `[]byte`
- 这类超限对象在 `Put` 时直接丢弃，不进入池

### 7.4 超大对象策略

当 `size > maxPooledCap` 时：

1. `Get` 直接分配精确大小的 `[]byte`。
2. 编码过程直接在这块返回的 `[]byte` 上进行。
3. `Put` 时直接丢弃，不回池。

## 8. API 行为语义

### 8.1 `Get(size int) []byte`

#### 规则

1. `size <= 0` 返回 `nil`。
2. 找到第一个 `bucket >= size`。
3. 命中 bucket：
   - 若池中有对象，返回 `buf[:size]`
   - 若无对象，新建 `cap == bucket` 的切片，再返回 `[:size]`
4. 超过 `maxPooledCap`：直接 `make([]byte, size)`，且该对象不参与池化回收。

### 8.2 `Put(buf []byte)`

#### 规则

1. `buf == nil`：忽略。
2. 只接受 `cap(buf)` **精确等于某个 bucket** 的对象。
3. 归还前长度重置为 0。
4. 若 `cap(buf)` 不匹配 bucket：直接丢弃。
5. 若超过 `maxPooledCap`：直接丢弃。

#### 关键原则

**不按 len 入池，只按 cap 入池。**

原因：

- 同一个 bucket 的复用语义必须稳定。
- 防止 `buf[:small]` 伪装成小对象入池。

### 8.3 `Buffer.Release()`

#### 规则

1. 只能释放一次。
2. 重复释放应在 debug 模式下可检测。
3. 释放后对象进入不可再用状态。

#### 设计目的

避免以下错误：

- 重复归还
- 归还后继续使用
- 同一对象被多个持有者共享后释放

### 8.4 `Scope.Close()`

#### 规则

1. 按登记顺序逐个归还。
2. 允许多次调用，但只有第一次生效。
3. 调用后 `Scope` 进入关闭状态。

## 9. ownership 与安全规则

这是整个设计中**最重要的部分**。

### 9.1 基本原则

谁借出，谁负责释放。

只有在**确认没有任何后续引用**时，才能归还到池。

### 9.2 原生 `[]byte` 的风险

直接使用原生 `[]byte` 时，风险最大：

1. slice 是引用语义。
2. 下游可能持有子切片。
3. 异步 goroutine 可能在调用返回后仍使用该数据。

因此：

- 热点路径可用原生 `[]byte`
- 复杂边界建议改用 `Buffer` 或 `Scope`

### 9.3 RocksDB / cgo 边界规则

只有在满足以下条件时，才允许在调用返回后立即归还 buffer：

1. C/Go 边界调用在返回前已经完成数据复制。
2. 没有异步队列继续持有这段内存。
3. 没有后台 batcher 延后消费这段数据。

否则必须：

1. 延迟释放。
2. 或者做一份 detached copy。

### 9.4 禁止场景

以下行为必须在文档和后续实现中明确禁止：

1. 把正在被异步使用的 `[]byte` / `Buffer` 放回池。
2. 把超大对象伪装成小对象入池。
3. 在 `Release` 之后继续持有 `Bytes()` 返回值并使用。
4. 跨多个 goroutine 同时写同一个 `Buffer`。

## 10. 与旧实现的关系

### 10.1 保留的部分

从旧实现中保留：

1. `sync.Pool` 作为底层复用机制。
2. 分桶管理思路。
3. 请求级统一释放思路（由 `Chunk` 演化为 `Scope`）。

### 10.2 放弃的部分

放弃旧实现中以下设计：

1. 一个 `ByteBuffer` 既是池对象又是编解码工具类。
2. 大量 `AppendUInt64`、`ReadInt32` 之类协议方法直接塞在池化对象上。
3. 模块边界过宽，导致池模块承担太多职责。

### 10.3 新设计的收益

1. 更清晰的模块边界。
2. 更适合作为多项目共享基础库。
3. 更易于后续扩展单独的 codec/buffer 工具模块。

## 11. 未来实现时的文件拆分建议

本节只描述未来建议，不代表现在就开始编码。

建议后续实现时拆分为：

```text
mempool/
  interface.go      // BytePool, StatsCollector
  options.go        // Options 与默认值
  bucket.go         // SizeClass 与 bucket 查找逻辑
  pool.go           // BucketedPool 主实现
  buffer.go         // Buffer 包装对象
  scope.go          // Scope 请求级批量释放
  metrics.go        // 可选统计钩子
```

拆分原则：

1. 底层池逻辑和包装对象分离。
2. `Scope` 独立，以便未来可选启用。
3. 统计能力独立，避免污染主路径。

## 12. 未来验证重点

后续编码阶段需要重点验证：

1. `Get/Put` 的 bucket 命中正确性。
2. 大对象不入池策略是否生效。
3. `Buffer.Release()` 的重复释放保护。
4. `Scope.Close()` 的统一释放行为。
5. 网络包解析 + RocksDB 写入路径下是否能安全释放。
6. benchmark 下 alloc rate、GC、P99 抖动是否下降。

## 13. 最终结论

新的内存池模块采用如下最终方案：

1. **底层使用 `sync.Pool` + 固定 bucket 的 `BytePool`**。
2. **上层提供可选 `Buffer` 包装对象，但不承载重型编解码职责**。
3. **保留请求级 `Scope`，统一释放一组借出的对象**。
4. **严格约束 ownership、超大对象策略、异步边界和 RocksDB/cgo 边界**。

这套方案兼顾：

- 高性能热点路径
- 多项目共享场景
- 更清晰的模块职责边界
- 后续扩展空间

相较于“纯 `[]byte` 池”或“万能 `ByteBuffer` 池”，该方案更适合作为 `basekit` 的正式基础模块设计。
