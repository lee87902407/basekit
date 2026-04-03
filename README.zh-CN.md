# Basekit

## 项目定位

`basekit` 用于沉淀各项目可复用的通用基础能力，例如网络协议解析、内存池缓存、通用工具组件与基础设施封装。

## 当前阶段

当前仓库完成了初始化骨架与文档规范建设，后续功能会按主模块逐步加入。

## 目录说明

- `README.md`：统一入口，负责汇总导航。
- `README.zh-CN.md`：中文主说明。
- `README.en.md`：英文说明。
- `AGENTS.md`：AI 与协作开发规范。
- `docs/examples/`：按功能拆分的示例文档。
- `examples/`：按功能拆分的示例代码。

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
- 功能示例文档：[`docs/examples/README.md`](./docs/examples/README.md)
- 示例代码约定：[`examples/README.md`](./examples/README.md)
- AI 协作规范：[`AGENTS.md`](./AGENTS.md)
