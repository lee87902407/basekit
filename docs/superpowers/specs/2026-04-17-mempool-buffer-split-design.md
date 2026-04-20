# mempool Buffer 拆分设计说明

## 1. 背景

当前 `/Users/yanjie/data/code/local/basekit/mempool/buffer.go` 中的 `Buffer` 同时承担了“可写构建阶段”和“只读消费阶段”两类职责，导致以下问题：

1. 类型边界不清晰，调用方无法从类型层面分辨当前对象是否还允许写入。
2. 旧模型中 `Release()` 为公开方法，生命周期既可由调用方手动控制，也可由 `Scope.Close()` 统一控制，ownership 容易模糊。
3. 默认非 `debug` 构建下，`Release()` 后再次使用会自动恢复为可用状态，这与“写阶段 -> 读阶段 -> 统一释放”的单向生命周期不一致。

本次改造目标是把 `Buffer` 拆成 `WriterBuffer` 与 `ReaderBuffer`，并把生命周期管理收拢到 `Scope`。

## 2. 目标

1. 删除公开 `Buffer` 类型与公开 `NewBuffer` 创建入口。
2. 新增 `WriterBuffer`，只承载写阶段能力。
3. 新增 `ReaderBuffer`，只承载只读阶段能力。
4. 由 `Scope.NewWriterBuffer(size int)` 作为唯一公开创建入口。
5. 提供 `WriterBuffer.ToReaderBuffer()`，把内容所有权从 writer 转交给 reader。
6. 删除公开 `Release()`，统一由 `Scope.Close()` 负责归还到底层 `BytePool`。
7. 新模型下，不再支持 release 后自动恢复；无论 `debug` 还是非 `debug`，失效对象都不可再继续使用。
8. 同步更新测试、README、示例与设计文档中对 `Buffer` 的描述。

## 3. 非目标

1. 不新增新的通用读写接口抽象。
2. 不修改 `BytePool`、bucket 策略、metrics 统计逻辑。
3. 不引入引用计数、共享读视图或 copy-on-write 等复杂 ownership 模型。
4. 不在本次改造中扩展 `ReaderBuffer` 的功能范围到“带游标读取器”之类新能力。

## 4. 总体设计

### 4.1 类型拆分

原有 `Buffer` 拆成两个公开类型：

- `WriterBuffer`
- `ReaderBuffer`

其中：

- `WriterBuffer` 用于构建内容，可读可写。
- `ReaderBuffer` 用于消费内容，只保留不会修改底层内容的只读方法。

### 4.2 生命周期模型

#### WriterBuffer

`WriterBuffer` 只有两种有效状态：

1. **writer-active**：刚创建后，可读可写。
2. **released**：发生以下任一事件后进入该状态：
   - 调用 `ToReaderBuffer()`，ownership 已转移给 `ReaderBuffer`
   - `Scope.Close()` 统一回收

进入 `released` 后，`WriterBuffer` 永久失效，不再允许任何读写或再次转换。

#### ReaderBuffer

`ReaderBuffer` 只有两种有效状态：

1. **reader-active**：由 `WriterBuffer.ToReaderBuffer()` 创建后，可只读访问。
2. **released**：`Scope.Close()` 统一回收后进入该状态。

进入 `released` 后，`ReaderBuffer` 永久失效，不再允许访问。

### 4.3 Scope 独占生命周期管理

`Scope` 成为 buffer 对象的唯一释放者：

- 创建：`Scope.NewWriterBuffer(size int) *WriterBuffer`
- 回收：`Scope.Close()`

调用方不能再手动调用公开 `Release()`。

## 5. API 设计

### 5.1 删除的 API

以下公开 API 删除：

1. `type Buffer struct`
2. `func NewBuffer(pool BytePool, size int) *Buffer`
3. `func (b *Buffer) Release()`

### 5.2 WriterBuffer API

`WriterBuffer` 保留以下公开方法：

1. `Bytes() []byte`
2. `Len() int`
3. `Cap() int`
4. `Released() bool`
5. `Reset()`
6. `Clone() []byte`
7. `DetachedCopy() []byte`
8. `EnsureCapacity(additional int)`
9. `Resize(n int)`
10. `Append(p []byte)`
11. `AppendByte(v byte)`
12. `ToReaderBuffer() *ReaderBuffer`

#### 5.2.1 ToReaderBuffer 语义

`ToReaderBuffer()` 的语义如下：

1. 调用时 `WriterBuffer` 必须处于 writer-active 状态。
2. 新建一个 `ReaderBuffer`。
3. 将 `WriterBuffer` 当前持有的 `buf` 与 `pool` 所有权转交给 `ReaderBuffer`。
4. 原 `WriterBuffer` 立即进入 `released` 状态。
5. 原 `WriterBuffer` 的 `buf` 清空，后续任何方法访问都视为对失效对象的访问。
6. 返回新的 `ReaderBuffer`。

### 5.3 ReaderBuffer API

`ReaderBuffer` 只保留只读能力：

1. `Bytes() []byte`
2. `Len() int`
3. `Cap() int`
4. `Released() bool`
5. `Clone() []byte`
6. `DetachedCopy() []byte`

不提供以下能力：

- `Reset()`
- `EnsureCapacity()`
- `Resize()`
- `Append()`
- `AppendByte()`
- `ToReaderBuffer()`

## 6. Scope 改造

### 6.1 结构调整

当前 `/Users/yanjie/data/code/local/basekit/mempool/scope.go` 中：

- `buffers []*Buffer`

将改为：

- `writers []*WriterBuffer`
- `readers []*ReaderBuffer`

`raws [][]byte` 保持不变。

### 6.2 创建入口

新增：

- `func (s *Scope) NewWriterBuffer(size int) *WriterBuffer`

删除：

- `func (s *Scope) NewBuffer(size int) *Buffer`

### 6.3 关闭逻辑

`Scope.Close()` 改造后逻辑：

1. 幂等检查保持不变。
2. 遍历 `writers`，仅回收尚未 `released` 的 writer。
3. 遍历 `readers`，仅回收尚未 `released` 的 reader。
4. 遍历 `raws`，归还裸 `[]byte`。
5. 标记 `closed = true`。

### 6.4 ownership 转移与 Scope 跟踪

`WriterBuffer.ToReaderBuffer()` 触发 ownership 转移后，必须让 `Scope` 也能追踪到新 reader，否则 `Scope.Close()` 会漏回收。

因此 `WriterBuffer` 需要持有对所属 `Scope` 的引用，或通过等效内部机制把新创建的 `ReaderBuffer` 注册到 `Scope.readers` 中。

本次设计采用：

- `WriterBuffer` 内部持有 `scope *Scope`
- `ToReaderBuffer()` 创建 `ReaderBuffer` 后，自动注册到该 `scope`

这样能确保：

1. `Scope.NewWriterBuffer()` 创建出的 writer 总是受 `Scope` 管理。
2. writer 转 reader 后，reader 也自动纳入同一 `Scope` 的释放范围。
3. `Scope.Close()` 不会重复回收同一块底层内存。

## 7. 文件拆分设计

建议把现有 `/Users/yanjie/data/code/local/basekit/mempool/buffer.go` 拆成以下文件：

1. `/Users/yanjie/data/code/local/basekit/mempool/writer_buffer.go`
2. `/Users/yanjie/data/code/local/basekit/mempool/reader_buffer.go`
3. `/Users/yanjie/data/code/local/basekit/mempool/writer_buffer_debug_checks.go`
4. `/Users/yanjie/data/code/local/basekit/mempool/writer_buffer_nodebug_checks.go`
5. `/Users/yanjie/data/code/local/basekit/mempool/reader_buffer_debug_checks.go`
6. `/Users/yanjie/data/code/local/basekit/mempool/reader_buffer_nodebug_checks.go`

这样做的原因：

- 把 writer / reader 的职责边界直接反映到文件层面。
- 保留现有 debug / non-debug 分离模式，减少与仓库现有风格的偏差。
- 让“writer 的可写校验”和“reader 的只读有效性校验”各自独立，不把状态逻辑重新揉回一个大文件里。

## 8. debug / non-debug 行为

### 8.1 与旧模型的差异

旧模型中，`/Users/yanjie/data/code/local/basekit/mempool/buffer_nodebug_checks.go` 会在非 `debug` 模式下把 `released` 的 `Buffer` 自动恢复为可用。这一行为与新设计冲突。

本次改造后：

- 无论 `debug` 还是非 `debug`
- 只要对象进入 `released`
- 就不允许继续使用，也不会自动恢复

### 8.2 debug 模式

在 `debug` 构建下：

- 对 released 的 `WriterBuffer` 做任何访问都 panic
- 对 released 的 `ReaderBuffer` 做任何访问都 panic
- 对已 released 的 `WriterBuffer` 再次调用 `ToReaderBuffer()` 也 panic

### 8.3 非 debug 模式

在非 `debug` 构建下，仍保持“失效对象不可继续使用”的约束。为避免静默错误，本次设计建议仍然对 released 对象访问直接 panic，而不是像旧模型那样自动恢复。

原因：

1. 新生命周期是单向的，自动恢复会破坏类型语义。
2. writer -> reader 的 ownership 转移后再恢复 writer，会产生双持有风险。
3. 这里的行为属于对象状态错误，而不是普通容错场景。

## 9. 内部实现约束

### 9.1 内部回收方法

虽然公开 `Release()` 删除，但 writer / reader 内部仍需要一个未导出的回收方法，例如：

- `releaseToPool()`

由 `Scope.Close()` 统一调用。

### 9.2 回收规则

writer 回收：

- 仅当 writer 仍持有底层 `buf` 且未 released 时回收。

reader 回收：

- 仅当 reader 持有底层 `buf` 且未 released 时回收。

这样可以确保 ownership 只会被归还一次。

### 9.3 Clone 与 DetachedCopy

`Clone()` 与 `DetachedCopy()` 继续保留“返回脱离 pool 生命周期影响的新切片副本”的语义。

## 10. 测试设计

### 10.1 WriterBuffer 测试

新增或迁移测试覆盖：

1. `Scope.NewWriterBuffer()` 能创建 writer。
2. `Append()` / `AppendByte()` 能正常写入。
3. `EnsureCapacity()` 仍通过 `BytePool` 扩容。
4. `Resize()` 仍通过 `BytePool` 扩容。
5. `Clone()` / `DetachedCopy()` 返回独立副本。
6. `ToReaderBuffer()` 会转交内容。
7. `ToReaderBuffer()` 后 writer 的 `Released()` 为 `true`。
8. `ToReaderBuffer()` 后 writer 再访问会失败。

### 10.2 ReaderBuffer 测试

新增测试覆盖：

1. `ReaderBuffer.Bytes()` / `Len()` / `Cap()` 返回正确内容。
2. `Clone()` / `DetachedCopy()` 返回独立副本。
3. `Scope.Close()` 后 reader 进入 released 状态。
4. reader 不具备写接口（由类型系统自然保证）。

### 10.3 debug 测试

重写原 `buffer_test.go` 中相关用例，覆盖：

1. writer 转 reader 后再访问 panic。
2. `Scope.Close()` 后 reader 再访问 panic。
3. writer 重复转换 panic。

### 10.4 非 debug 测试

替换原 `/Users/yanjie/data/code/local/basekit/mempool/buffer_nodebug_test.go` 中“release 后恢复”的断言，改为覆盖：

1. writer 转 reader 后再访问仍失败。
2. `Scope.Close()` 后 reader 再访问仍失败。
3. `Scope.Close()` 不重复归还同一块内存。

## 11. 文档与示例更新

由于这是主模块公开 API 变化，需同步更新：

1. `/Users/yanjie/data/code/local/basekit/README.md`
2. `/Users/yanjie/data/code/local/basekit/README.zh-CN.md`
3. `/Users/yanjie/data/code/local/basekit/README.en.md`
4. `/Users/yanjie/data/code/local/basekit/docs/examples/mempool.md`
5. `/Users/yanjie/data/code/local/basekit/examples/mempool/main.go`
6. `/Users/yanjie/data/code/local/basekit/docs/designs/memory-pool-design.md`

文档中需要明确说明：

- `Scope.NewWriterBuffer()` 是创建入口。
- `WriterBuffer` 与 `ReaderBuffer` 的职责区别。
- `ToReaderBuffer()` 会转移 ownership。
- `WriterBuffer` 转换后立即失效。
- `Scope.Close()` 是统一释放入口。

## 12. 风险与取舍

### 12.1 API 破坏性变更

本次是显式破坏性调整，调用方需要从：

- `Buffer`
- `NewBuffer`
- 公开 `Release()`

迁移到：

- `WriterBuffer`
- `ReaderBuffer`
- `Scope.NewWriterBuffer()`
- `Scope.Close()`

这是有意为之，用类型边界换取更清晰的 ownership 语义。

### 12.2 非 debug 模式更严格

旧行为更宽松，新行为更严格，可能暴露以前未显现的错误用法。但这与“单向生命周期”设计一致，值得接受。

### 12.3 Scope 与 buffer 耦合增强

`WriterBuffer` 需要知道所属 `Scope` 才能在 `ToReaderBuffer()` 时注册 reader。这会增加少量内部耦合，但相比漏回收或双回收风险，这是更合理的取舍。

## 13. 验证要求

实现完成后至少需要验证：

1. `/Users/yanjie/data/code/local/basekit/mempool` 相关 LSP diagnostics 无错误。
2. `go test ./...` 通过。
3. `go test -tags debug ./...` 通过。
4. README、示例文档、示例代码中的 `Buffer` 入口已更新为新模型。
5. 不存在公开 `NewBuffer` 与公开 `Release()` 的残留出口。
