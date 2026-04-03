# Basekit

## 项目定位

`basekit` 用于沉淀各项目可复用的通用基础能力，例如网络协议解析、内存池缓存、通用工具组件与基础设施封装。

## 当前阶段

当前仓库已落地首个正式主模块 `mempool`，后续其他基础能力会按主模块逐步加入。

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
- `Buffer` 包装对象
- `Scope` 请求级统一释放
- `debug` 构建标签下的 `Buffer` 误用检查

相关入口：

- 设计文档：[`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- 示例文档：[`docs/examples/mempool.md`](./docs/examples/mempool.md)
- 示例代码：[`examples/mempool/`](./examples/mempool/)

行为说明：

- 默认构建下，`Buffer` 不会对 `use after release` 或重复 `Release` 做 panic 检查；如果释放后又继续使用，会自动恢复为可继续托管、可被后续 `Scope.Close()` 回收的状态。
- 使用 `go test -tags debug ./...`、`go build -tags debug ./...` 等方式加入 `debug` 标签时，会启用 `Buffer` 的运行时安全检查，用于在开发和测试阶段更早暴露误用。

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
- 设计文档：[`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)
- 功能示例文档：[`docs/examples/README.md`](./docs/examples/README.md)
- 示例代码约定：[`examples/README.md`](./examples/README.md)
- AI 协作规范：[`AGENTS.md`](./AGENTS.md)
