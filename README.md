# Basekit

通用基础库项目统一入口。

## 文档入口

- 中文说明：[`README.zh-CN.md`](./README.zh-CN.md)
- English Guide: [`README.en.md`](./README.en.md)
- 功能示例文档入口：[`docs/examples/README.md`](./docs/examples/README.md)
- 示例代码约定：[`examples/README.md`](./examples/README.md)
- AI 协作规范：[`AGENTS.md`](./AGENTS.md)
- 内存池设计文档：[`docs/designs/memory-pool-design.md`](./docs/designs/memory-pool-design.md)

## 当前模块

- `mempool`：基于 `sync.Pool` 的 `[]byte` 分桶内存池，支持 `Buffer` 包装、`Scope` 请求级批量释放，以及可选的 `debug` 构建期安全检查。

## 模块文档与示例

- `mempool` 示例文档：[`docs/examples/mempool.md`](./docs/examples/mempool.md)
- `mempool` 示例代码：[`examples/mempool/`](./examples/mempool/)

后续每新增一个主模块时，必须同步完成以下事项：

1. 更新本入口 README 与中英文说明文档。
2. 新增对应的功能示例文档：`docs/examples/<module>.md`。
3. 新增或更新对应的示例代码目录：`examples/<module>/`。
4. 在 README 中补充跳转链接，保持文档入口可发现。
