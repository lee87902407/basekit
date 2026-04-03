# log 模块设计方案

## 目标

新增 `log` 主模块，基于 `github.com/uber-go/zap` 提供统一日志封装，满足以下目标：

1. 提供全局单例初始化入口 `log.Init(Config) error`。
2. 业务使用时只需要调用包级方法，如：
   - `log.Debug(msg string, kv ...any)`
   - `log.Info(msg string, kv ...any)`
   - `log.Warn(msg string, kv ...any)`
   - `log.Error(msg string, kv ...any)`
   - `log.Sync() error`
3. 初始化参数覆盖以下能力：
   - `logPath string`
   - `logName string`
   - `level Level`
   - `maxSizeMB uint16`
   - `maxBackup uint16`
   - `maxAge uint16`
   - `compress bool`
   - `output OutputMode`
4. 支持控制台输出、文件输出、控制台与文件同时输出。
5. 控制台输出采用易读文本格式，文件输出采用 JSON 格式。
6. 文件输出支持按大小轮转与历史文件保留。

## 依赖选择

### 核心日志库

`log` 模块使用 `go.uber.org/zap` 作为底层日志实现，原因如下：

1. 提供结构化日志能力。
2. 性能稳定，适合作为基础库默认日志实现。
3. 支持 `zap.Logger` 与 `zap.SugaredLogger` 双层能力。
4. 支持 `zapcore.NewTee` 组合多个输出目标。
5. 支持 `zap.AtomicLevel` 实现运行时级别切换。

### 文件轮转库

文件输出使用 `gopkg.in/natefinch/lumberjack.v2`，原因如下：

1. 是 zap 官方 FAQ 明确推荐的轮转方案。
2. 参数模型与本模块需求高度一致：
   - `MaxSize`
   - `MaxBackups`
   - `MaxAge`
   - `Compress`
3. 与 `zapcore.AddSync` 配合简单，不需要额外复杂桥接层。

## 模块定位

`log` 是基础日志模块，职责范围控制如下：

### 当前模块职责

1. 提供统一初始化接口。
2. 提供包级单例日志调用入口。
3. 提供结构化键值日志写入能力。
4. 提供控制台 / 文件 / 双输出模式。
5. 提供运行时日志级别调整入口。
6. 提供底层 `*zap.Logger` 与 `*zap.SugaredLogger` 的高级访问入口。

### 当前模块非目标

以下内容明确不纳入当前版本：

1. 不接管 zap 全局 logger（不默认调用 `zap.ReplaceGlobals`）。
2. 不提供分布式 trace/span 自动注入。
3. 不提供异步日志缓冲队列。
4. 不提供多文件路由（例如 info/error 分不同文件）。
5. 不提供自定义 encoder 插件扩展系统。
6. 不提供配置热重载。

## 核心类型设计

### Level

模块定义自己的 `Level` 类型，对外屏蔽 zap 细节，建议支持：

1. `LevelDebug`
2. `LevelInfo`
3. `LevelWarn`
4. `LevelError`

设计目的：

1. 保持 `log` 模块对外 API 稳定。
2. 避免业务侧直接依赖 zap 的 level 枚举。
3. 便于后续增加字符串解析与配置映射。

### OutputMode

模块定义输出模式枚举：

1. `OutputModeConsole`
2. `OutputModeFile`
3. `OutputModeBoth`

设计目的：

1. 精确表达控制台 / 文件 / 双输出三种状态。
2. 避免使用多个布尔值带来的组合歧义。

### Config

模块对外暴露平铺配置结构：

1. `LogPath string`
2. `LogName string`
3. `Level Level`
4. `MaxSizeMB uint16`
5. `MaxBackup uint16`
6. `MaxAge uint16`
7. `Compress bool`
8. `Output OutputMode`

约束说明：

1. 当 `Output` 包含文件输出时，`LogPath` 与 `LogName` 必须可组成有效文件路径。
2. 当 `Output` 为 `OutputModeConsole` 时，可忽略文件轮转参数。
3. `MaxSizeMB`、`MaxBackup`、`MaxAge` 使用 `uint16`，保持与用户要求一致。

## 初始化语义

### Init 行为

`Init(Config) error` 的行为定义如下：

1. 只允许成功初始化一次。
2. 第二次调用直接返回错误，不覆盖旧配置。
3. 初始化成功后，包级 `Debug/Info/Warn/Error` 均转发到正式 logger。

### Init 前的兜底行为

如果业务在 `Init` 之前调用：

1. `log.Debug(...)`
2. `log.Info(...)`
3. `log.Warn(...)`
4. `log.Error(...)`

则使用一个默认控制台 logger 顶上，而不是静默丢弃。

该兜底 logger 的设计意图如下：

1. 避免应用启动早期日志丢失。
2. 保持业务侧调用方式一致，不要求手动判空。
3. 将“未初始化”状态的损害控制在输出能力降级，而不是功能不可用。

兜底 logger 的行为边界：

1. 仅输出到控制台。
2. 使用文本格式。
3. 不写入文件。
4. 不参与正式配置中的轮转能力。
5. `Sync()` 在未初始化阶段仅同步兜底 logger 对应输出。

## 输出设计

### 控制台输出

控制台输出使用 `zapcore.NewConsoleEncoder`，原因如下：

1. 本地开发更易读。
2. 与文件 JSON 输出形成明确分工。

控制台输出目标使用标准输出流，并通过 zap 的同步包装保证并发写入安全。

### 文件输出

文件输出使用 `zapcore.NewJSONEncoder`，原因如下：

1. 便于日志采集系统消费。
2. 便于结构化检索与后续机器分析。

文件 sink 通过 `lumberjack.Logger` 构建，并由 `zapcore.AddSync` 接入 zap core。

### 双输出

当 `Output` 为 `OutputModeBoth` 时：

1. 控制台 core 使用文本 encoder。
2. 文件 core 使用 JSON encoder。
3. 两个 core 通过 `zapcore.NewTee` 合并。

## 调用接口设计

### 包级便捷方法

模块对外暴露以下包级方法：

1. `Debug(msg string, kv ...any)`
2. `Info(msg string, kv ...any)`
3. `Warn(msg string, kv ...any)`
4. `Error(msg string, kv ...any)`
5. `Sync() error`

键值参数采用 `kv ...any`，内部转换为 zap 可接受的字段形式。

### 高级访问入口

除包级便捷方法外，还提供：

1. `L() *zap.Logger`
2. `S() *zap.SugaredLogger`
3. `SetLevel(level Level) error`

设计目的：

1. 满足复杂场景直接操作 zap logger 的需求。
2. 保留运行时动态调级能力。
3. 不破坏“默认简单使用”的主路径。

## 级别控制设计

模块内部使用 `zap.AtomicLevel` 保存当前级别状态。

带来的能力：

1. 初始化时按 `Config.Level` 设置初始级别。
2. 运行中通过 `SetLevel` 动态修改级别。
3. 控制台与文件输出共享同一套级别门限。

## 文件路径设计

文件输出路径由以下两部分组成：

1. `LogPath`
2. `LogName`

实际日志文件路径为：

- `filepath.Join(LogPath, LogName)`

设计要求：

1. 不在 API 中额外要求业务自己拼完整文件名。
2. 输出到文件时，模块内部统一负责路径组合。

## 错误处理设计

### Init 可能返回的错误

1. 重复初始化。
2. 文件输出模式下配置无效，例如缺失必要文件路径信息。
3. 级别或输出模式非法。
4. 底层 logger 构建失败。

### 非 Init 路径的错误处理

1. `Debug/Info/Warn/Error` 不向业务返回错误。
2. `Sync()` 返回错误，供业务在退出前处理。
3. `SetLevel()` 返回错误，避免非法级别修改被静默忽略。

## 并发与状态管理

模块内部维护一套包级状态，至少包括：

1. 正式 logger
2. sugared logger
3. fallback 控制台 logger
4. atomic level
5. 初始化状态

设计要求：

1. 并发读日志调用必须安全。
2. 初始化判定必须安全。
3. 不允许出现半初始化状态被业务观察到。

## 文档与示例要求

由于 `log` 是新增主模块，同一次变更必须同步完成：

1. `README.md` 增加 `log` 模块入口。
2. `README.zh-CN.md` 增加 `log` 模块中文说明。
3. `README.en.md` 增加 `log` 模块英文说明。
4. 新增 `docs/examples/log.md`。
5. 新增 `examples/log/main.go`。
6. 更新 `docs/examples/README.md` 与 `examples/README.md` 索引。

## 测试设计方向

实现阶段至少需要覆盖以下测试：

1. `Init` 首次初始化成功。
2. 重复 `Init` 返回错误。
3. 未初始化时调用 `Info/Debug` 不 panic，且走 fallback 控制台 logger。
4. `OutputModeConsole` 仅输出控制台。
5. `OutputModeFile` 仅输出文件。
6. `OutputModeBoth` 同时输出控制台与文件。
7. `SetLevel` 能动态修改日志级别。
8. `Sync` 在不同状态下行为正确。

## 后续演进方向

如果未来日志需求继续扩展，可考虑在后续版本评估：

1. 显式 `Rotate()` 接口。
2. 可选接管 zap globals 的能力。
3. error 单独输出到 stderr。
4. 自定义时间格式与字段键名。

当前版本不主动纳入上述能力，以保持模块职责清晰。
