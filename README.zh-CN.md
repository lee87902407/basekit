# Basekit

## 项目定位

`basekit` 用于沉淀各项目可复用的通用基础能力，例如网络协议解析、内存池缓存、通用工具组件与基础设施封装。

## 当前阶段

当前仓库已落地正式主模块 `mempool` 与 `log`，后续其他基础能力会按主模块逐步加入。

## 目录说明

- `README.md`：统一入口，负责汇总导航。
- `README.zh-CN.md`：中文主说明。
- `README.en.md`：英文说明。
- `AGENTS.md`：AI 与协作开发规范。
- `docs/designs/`：正式设计文档。
- `docs/examples/`：按功能拆分的示例文档。
- `examples/`：按功能拆分的示例代码。

## 当前模块

### mempool

`mempool` 是一个基于 `sync.Pool` 的 `[]byte` 分桶内存池模块，适用于高频短生命周期缓冲区场景，支持：

- 512KB 以内 bucket 化复用
- 超限对象直接分配、归还丢弃
- 统一的 `Buffer` 接口
- `Buffer.Type()` 类型标识与 `BufferTypeHeap` / `BufferTypeNative` 常量
- 可写的 `HeapBuffer` 包装对象
- cgo 场景下只读的 `NativeBuffer`
- `Scope` 请求级统一释放
- `debug` 构建标签下的 `Buffer` 误用检查

相关入口：

- 设计文档：[`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- 示例文档：[`docs/examples/mempool.md`](./docs/examples/mempool.md)
- 示例代码：[`examples/mempool/`](./examples/mempool/)

行为说明：

- `mempool.NewHeapBuffer` 返回池化可写缓冲；`Scope.NewBuffer` 继续返回 `*HeapBuffer`，并由 `HeapBuffer` 实现统一的 `Buffer` 接口。
- cgo 开启时可通过 `mempool.NewNativeBuffer`、`Scope.NewNativeBuffer` 或 `Scope.GetNativeBuffer` 包装原生内存；`NativeBuffer` 只支持只读访问与释放，任何写入类方法都会直接 panic，释放时会调用注入的 `freeFun`。
- `Scope.GetHeapBuffer` 用于申请并托管原始 heap `[]byte`；原 `Track` 已移除，外部原始切片不再单独挂入 `Scope`。
- 默认构建下，`HeapBuffer` 不会对 `use after release` 或重复 `Release` 做 panic 检查；如果释放后又继续使用，会自动恢复为可继续托管、可被后续 `Scope.Close()` 回收的状态。
- 使用 `go test -tags debug ./...`、`go build -tags debug ./...` 等方式加入 `debug` 标签时，会启用 `Buffer` 生命周期的运行时安全检查，用于在开发和测试阶段更早暴露误用。
- `Scope` 一旦 `Close()`，后续再调用 `GetHeapBuffer`、`NewBuffer`、`NewNativeBuffer` 或 `GetNativeBuffer` 都会直接 panic，避免新增资源脱离回收路径。

### log

`log` 是一个基于 `zap` 的全局单例日志模块，适用于服务启动日志、业务结构化日志和本地开发调试场景，支持：

- 全局 `Init(Config)` 初始化
- `Debug` / `Info` / `Warn` / `Error` / `Sync` 包级调用
- `OutputModeConsole` / `OutputModeFile` / `OutputModeBoth`
- 控制台文本输出与文件 JSON 输出
- 基于 `lumberjack` 的按大小轮转与历史保留
- 运行时通过 `SetLevel` 动态调级
- 未初始化阶段使用默认控制台 logger 兜底

相关入口：

- 设计文档：[`docs/designs/log-design.md`](./docs/designs/log-design.md)
- 示例文档：[`docs/examples/log.md`](./docs/examples/log.md)
- 示例代码：[`examples/log/`](./examples/log/)

行为说明：

- `Init` 只允许成功一次，重复调用会返回错误。
- 如果业务在 `Init` 前调用 `Debug` / `Info` / `Warn` / `Error`，模块会退化为默认控制台 logger，而不是直接丢日志。
- 控制台输出使用文本格式，文件输出使用 JSON 格式；双输出模式下二者同时生效。

## 文档维护规则

新增一个主模块时，必须同时完成以下更新：

1. 更新 `README.md`，补充模块简介与链接。
2. 更新本文档与 `README.en.md`，保持中英文入口一致。
3. 新增该模块的示例文档：`docs/examples/<module>.md`。
4. 新增或更新该模块的示例代码：`examples/<module>/`。
5. 如模块接口、行为或用法变化，必须同步修正文档与示例。

## 规范说明

- 仓库中的注释、说明性文档、开发规范默认使用中文编写。
- 对外展示或跨团队协作需要英文时，同时维护英文版本。
- 示例文档应聚焦“模块做什么、何时使用、如何接入、最小示例、注意事项”。

## 相关入口

- 统一入口：[`README.md`](./README.md)
- 日志模块设计文档：[`docs/designs/log-design.md`](./docs/designs/log-design.md)
- 设计文档：[`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- 功能示例文档：[`docs/examples/README.md`](./docs/examples/README.md)
- 示例代码约定：[`examples/README.md`](./examples/README.md)
- AI 协作规范：[`AGENTS.md`](./AGENTS.md)
