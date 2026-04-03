# Log Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 `log` 主模块，基于 zap 与 lumberjack 提供全局单例日志封装、控制台/文件/双输出、运行时调级、未初始化兜底控制台 logger，并同步补齐文档与示例。

**Architecture:** `log` 包内部维护正式 logger、sugared logger、fallback 控制台 logger 与 atomic level。初始化使用 `Config + OutputMode + Level` 建立 core；控制台使用 ConsoleEncoder，文件使用 JSONEncoder，双输出通过 `zapcore.NewTee` 合并；业务侧统一通过包级 `Debug/Info/Warn/Error/Sync` 访问。

**Tech Stack:** Go 1.25、`go.uber.org/zap`、`gopkg.in/natefinch/lumberjack.v2`

---

### Task 1: 建立配置与类型骨架

**Files:**
- Create: `log/options.go`
- Test: `log/options_test.go`

- [ ] **Step 1: Write the failing test**

```go
package log

import "testing"

func TestDefaultConfigProvidesUsableDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != LevelInfo {
		t.Fatalf("cfg.Level = %v, want %v", cfg.Level, LevelInfo)
	}
	if cfg.Output != OutputModeConsole {
		t.Fatalf("cfg.Output = %v, want %v", cfg.Output, OutputModeConsole)
	}
	if cfg.MaxSizeMB == 0 {
		t.Fatal("cfg.MaxSizeMB should not be zero")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./log -run TestDefaultConfigProvidesUsableDefaults`
Expected: FAIL with undefined `DefaultConfig` / `LevelInfo` / `OutputModeConsole`

- [ ] **Step 3: Write minimal implementation**

```go
package log

type Level uint8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type OutputMode uint8

const (
	OutputModeConsole OutputMode = iota
	OutputModeFile
	OutputModeBoth
)

type Config struct {
	LogPath   string
	LogName   string
	Level     Level
	MaxSizeMB uint16
	MaxBackup uint16
	MaxAge    uint16
	Compress  bool
	Output    OutputMode
}

func DefaultConfig() Config {
	return Config{
		Level:     LevelInfo,
		MaxSizeMB: 100,
		MaxBackup: 10,
		MaxAge:    30,
		Compress:  true,
		Output:    OutputModeConsole,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./log -run TestDefaultConfigProvidesUsableDefaults`
Expected: PASS

### Task 2: 先建立未初始化 fallback 行为

**Files:**
- Create: `log/logger.go`
- Create: `log/logger_test.go`
- Modify: `log/options.go`

- [ ] **Step 1: Write the failing test**

```go
package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestInfoUsesFallbackConsoleLoggerBeforeInit(t *testing.T) {
	resetForTest()
	buf := &bytes.Buffer{}
	setFallbackOutputForTest(buf)

	Info("booting", "stage", "pre-init")

	text := buf.String()
	if !strings.Contains(text, "booting") {
		t.Fatalf("fallback output = %q, want contains booting", text)
	}
	if !strings.Contains(text, "pre-init") {
		t.Fatalf("fallback output = %q, want contains pre-init", text)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./log -run TestInfoUsesFallbackConsoleLoggerBeforeInit`
Expected: FAIL with undefined `Info` / helper functions

- [ ] **Step 3: Write minimal implementation**

```go
package log

import (
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	mu            sync.RWMutex
	logger        *zap.Logger
	sugar         *zap.SugaredLogger
	fallbackOnce  sync.Once
	fallbackSugar *zap.SugaredLogger
	fallbackOut   io.Writer = os.Stdout
)

func fallbackLogger() *zap.SugaredLogger {
	fallbackOnce.Do(func() {
		enc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core := zapcore.NewCore(enc, zapcore.AddSync(fallbackOut), zap.DebugLevel)
		fallbackSugar = zap.New(core, zap.AddCallerSkip(1)).Sugar()
	})
	return fallbackSugar
}

func activeSugar() *zap.SugaredLogger {
	mu.RLock()
	defer mu.RUnlock()
	if sugar != nil {
		return sugar
	}
	return fallbackLogger()
}

func Debug(msg string, kv ...any) { activeSugar().Debugw(msg, kv...) }
func Info(msg string, kv ...any)  { activeSugar().Infow(msg, kv...) }
func Warn(msg string, kv ...any)  { activeSugar().Warnw(msg, kv...) }
func Error(msg string, kv ...any) { activeSugar().Errorw(msg, kv...) }

func resetForTest() {
	mu.Lock()
	defer mu.Unlock()
	logger = nil
	sugar = nil
	fallbackSugar = nil
	fallbackOnce = sync.Once{}
	fallbackOut = os.Stdout
}

func setFallbackOutputForTest(w io.Writer) {
	fallbackSugar = nil
	fallbackOnce = sync.Once{}
	fallbackOut = w
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./log -run TestInfoUsesFallbackConsoleLoggerBeforeInit`
Expected: PASS

### Task 3: 实现 Init、输出模式与 Sync

**Files:**
- Modify: `log/logger.go`
- Modify: `log/logger_test.go`

- [ ] **Step 1: Write the failing test**

```go
package log

import (
	"path/filepath"
	"testing"
)

func TestInitRejectsSecondInitialization(t *testing.T) {
	resetForTest()
	cfg := DefaultConfig()
	if err := Init(cfg); err != nil {
		t.Fatalf("first init error = %v", err)
	}
	if err := Init(cfg); err == nil {
		t.Fatal("second init should return error")
	}
}

func TestInitWritesFileWhenOutputModeFile(t *testing.T) {
	resetForTest()
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Output = OutputModeFile
	cfg.LogPath = dir
	cfg.LogName = "app.log"

	if err := Init(cfg); err != nil {
		t.Fatalf("Init error = %v", err)
	}
	Info("hello-file", "k", "v")
	if err := Sync(); err != nil {
		t.Fatalf("Sync error = %v", err)
	}

	path := filepath.Join(dir, "app.log")
	assertFileContains(t, path, "hello-file")
	assertFileContains(t, path, "\"k\":\"v\"")
	resetForTest()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./log -run 'TestInitRejectsSecondInitialization|TestInitWritesFileWhenOutputModeFile'`
Expected: FAIL because `Init` / `Sync` / file output are not implemented

- [ ] **Step 3: Write minimal implementation**

实现要求：

1. 在 `logger.go` 中新增 `Init(cfg Config) error`。
2. 使用 `zap.AtomicLevel` 保存正式 logger 级别。
3. `OutputModeConsole`：仅创建 console core。
4. `OutputModeFile`：仅创建 file core，文件路径为 `filepath.Join(cfg.LogPath, cfg.LogName)`。
5. `OutputModeBoth`：使用 `zapcore.NewTee(consoleCore, fileCore)`。
6. 文件输出通过 `lumberjack.Logger` 接入。
7. 第二次 `Init` 返回错误。
8. `Sync()`：已初始化时同步正式 logger，否则同步 fallback logger。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./log -run 'TestInitRejectsSecondInitialization|TestInitWritesFileWhenOutputModeFile'`
Expected: PASS

### Task 4: 实现双输出、动态调级与高级访问

**Files:**
- Modify: `log/logger.go`
- Modify: `log/logger_test.go`

- [ ] **Step 1: Write the failing test**

```go
package log

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestOutputModeBothWritesConsoleAndFile(t *testing.T) {
	resetForTest()
	dir := t.TempDir()
	console := &bytes.Buffer{}
	setConsoleOutputForTest(console)

	cfg := DefaultConfig()
	cfg.Output = OutputModeBoth
	cfg.LogPath = dir
	cfg.LogName = "both.log"

	if err := Init(cfg); err != nil {
		t.Fatalf("Init error = %v", err)
	}
	Warn("dual", "mode", "both")
	if err := Sync(); err != nil {
		t.Fatalf("Sync error = %v", err)
	}

	if got := console.String(); got == "" {
		t.Fatal("console output should not be empty")
	}
	assertFileContains(t, filepath.Join(dir, "both.log"), "dual")
	resetForTest()
}

func TestSetLevelChangesActiveLevel(t *testing.T) {
	resetForTest()
	console := &bytes.Buffer{}
	setConsoleOutputForTest(console)

	cfg := DefaultConfig()
	cfg.Output = OutputModeConsole
	cfg.Level = LevelWarn

	if err := Init(cfg); err != nil {
		t.Fatalf("Init error = %v", err)
	}
	Info("suppressed")
	if console.Len() != 0 {
		t.Fatalf("console output = %q, want empty before level change", console.String())
	}
	if err := SetLevel(LevelInfo); err != nil {
		t.Fatalf("SetLevel error = %v", err)
	}
	Info("visible")
	if got := console.String(); got == "" {
		t.Fatal("console output should contain visible log")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./log -run 'TestOutputModeBothWritesConsoleAndFile|TestSetLevelChangesActiveLevel'`
Expected: FAIL because dual output or dynamic level helpers are incomplete

- [ ] **Step 3: Write minimal implementation**

实现要求：

1. 提供 `L() *zap.Logger`、`S() *zap.SugaredLogger`。
2. 提供 `SetLevel(level Level) error`。
3. console/file 共用一个 `zap.AtomicLevel`。
4. 测试辅助中支持替换正式 console 输出 writer。
5. 确保 `OutputModeBoth` 同时写入 console 与 file。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./log -run 'TestOutputModeBothWritesConsoleAndFile|TestSetLevelChangesActiveLevel'`
Expected: PASS

### Task 5: 补全 README、示例文档与示例代码

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `README.en.md`
- Modify: `docs/examples/README.md`
- Modify: `examples/README.md`
- Create: `docs/examples/log.md`
- Create: `examples/log/main.go`

- [ ] **Step 1: Write the failing test**

```go
package log

import "testing"

func TestSyncWorksWithFallbackLoggerBeforeInit(t *testing.T) {
	resetForTest()
	if err := Sync(); err != nil {
		t.Fatalf("Sync error = %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./log -run TestSyncWorksWithFallbackLoggerBeforeInit`
Expected: FAIL if fallback sync path is incomplete

- [ ] **Step 3: Write minimal implementation**

实现要求：

1. 让 fallback logger 的 `Sync()` 路径稳定可调用。
2. 更新三个 README，把 `log` 作为新主模块加入。
3. 新增 `docs/examples/log.md`，包含模块用途、场景、最小示例、注意事项、README 跳转关系。
4. 新增 `examples/log/main.go`，展示 `Init + Info + Sync` 最小用法。
5. 更新 `docs/examples/README.md` 与 `examples/README.md` 索引。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./log -run TestSyncWorksWithFallbackLoggerBeforeInit`
Expected: PASS

### Task 6: 完整验证与收尾

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Verify: `log/*.go`
- Verify: `docs/designs/log-design.md`

- [ ] **Step 1: Add dependencies**

Run: `go get go.uber.org/zap gopkg.in/natefinch/lumberjack.v2`
Expected: `go.mod` / `go.sum` updated

- [ ] **Step 2: Run focused tests**

Run: `go test ./log`
Expected: PASS

- [ ] **Step 3: Run full verification**

Run: `go test ./... && go run ./examples/log`
Expected: PASS

- [ ] **Step 4: Run diagnostics**

Run: language-server diagnostics for `log/*.go`
Expected: 0 errors

- [ ] **Step 5: Prepare review and push**

Run:

```bash
git status --short
git diff --stat
```

Expected: only intended `log` module, docs, examples, plan/design, and dependency changes remain
