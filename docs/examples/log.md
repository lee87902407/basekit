# log 示例文档

## 模块用途

`log` 提供一个基于 `zap` 的全局单例日志封装，用于统一业务日志输出、日志轮转策略和控制台 / 文件双通道输出方式。

## 典型使用场景

1. 服务启动与关闭日志。
2. 业务流程中的结构化键值日志。
3. 本地开发阶段的控制台调试输出。
4. 需要同时写控制台和文件的服务端进程。

## 最小可运行示例

```go
package main

import (
	"path/filepath"

	basekitlog "github.com/lee87902407/basekit/log"
)

func main() {
	cfg := basekitlog.DefaultConfig()
	cfg.Output = basekitlog.OutputModeBoth
	cfg.LogPath = filepath.Join(".", "tmp")
	cfg.LogName = "app.log"

	if err := basekitlog.Init(cfg); err != nil {
		panic(err)
	}

	basekitlog.Info("service started", "module", "log", "mode", "both")

	if err := basekitlog.Sync(); err != nil {
		panic(err)
	}
}
```

## 接入注意事项

1. `Init` 只允许成功一次，重复调用会返回错误。
2. `OutputModeFile` 与 `OutputModeBoth` 下必须提供有效的 `LogPath` 与 `LogName`。
3. 控制台输出为文本格式，文件输出为 JSON 格式。
4. 如果在 `Init` 之前调用 `Debug` / `Info` / `Warn` / `Error`，模块会使用默认控制台 logger 兜底。
5. 文件轮转依赖 `lumberjack`，`MaxSizeMB` / `MaxBackup` / `MaxAge` / `Compress` 仅在文件输出模式下生效。
6. 退出进程前建议调用 `Sync()`，以便刷新底层缓冲。

## 与 README 的跳转关系

1. 统一入口位于 [`README.md`](../../README.md)。
2. 中文说明位于 [`README.zh-CN.md`](../../README.zh-CN.md)。
3. 英文说明位于 [`README.en.md`](../../README.en.md)。
4. 对应示例代码位于 [`examples/log/`](../../examples/log/)。
