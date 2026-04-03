package log

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	mu            sync.RWMutex
	logger        *zap.Logger
	sugar         *zap.SugaredLogger
	levelCtl      zap.AtomicLevel
	inited        bool
	fallbackOnce  sync.Once
	fallbackSugar *zap.SugaredLogger
	fallbackOut   io.Writer = os.Stdout
	consoleOut    io.Writer = os.Stdout
)

var errAlreadyInitialized = errors.New("log: already initialized")

var errInvalidLevel = errors.New("log: invalid level")

func Init(cfg Config) error {
	mu.Lock()
	defer mu.Unlock()
	if inited {
		return errAlreadyInitialized
	}
	core, lvl, err := buildCore(cfg)
	if err != nil {
		return err
	}
	levelCtl = lvl
	logger = zap.New(core, zap.AddCallerSkip(1))
	sugar = logger.Sugar()
	inited = true
	return nil
}

func buildCore(cfg Config) (zapcore.Core, zap.AtomicLevel, error) {
	zapLevel, err := validateLevel(cfg.Level)
	if err != nil {
		return nil, zap.NewAtomicLevelAt(zap.InfoLevel), err
	}
	lvl := zap.NewAtomicLevelAt(zapLevel)
	switch cfg.Output {
	case OutputModeConsole:
		return newConsoleCore(lvl), lvl, nil
	case OutputModeFile:
		fileCore, err := newFileCore(cfg, lvl)
		return fileCore, lvl, err
	case OutputModeBoth:
		fileCore, err := newFileCore(cfg, lvl)
		if err != nil {
			return nil, lvl, err
		}
		return zapcore.NewTee(newConsoleCore(lvl), fileCore), lvl, nil
	default:
		return nil, lvl, errors.New("log: invalid output mode")
	}
}

func newConsoleCore(lvl zap.AtomicLevel) zapcore.Core {
	enc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	return zapcore.NewCore(enc, zapcore.AddSync(consoleOut), lvl)
}

func newFileCore(cfg Config, lvl zap.AtomicLevel) (zapcore.Core, error) {
	if cfg.LogPath == "" || cfg.LogName == "" {
		return nil, errors.New("log: file output requires log path and log name")
	}
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	writer := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogPath, cfg.LogName),
		MaxSize:    int(cfg.MaxSizeMB),
		MaxBackups: int(cfg.MaxBackup),
		MaxAge:     int(cfg.MaxAge),
		Compress:   cfg.Compress,
	}
	return zapcore.NewCore(enc, zapcore.AddSync(writer), lvl), nil
}

func validateLevel(level Level) (zapcore.Level, error) {
	switch level {
	case LevelDebug:
		return zap.DebugLevel, nil
	case LevelInfo:
		return zap.InfoLevel, nil
	case LevelWarn:
		return zap.WarnLevel, nil
	case LevelError:
		return zap.ErrorLevel, nil
	default:
		return zap.InfoLevel, errInvalidLevel
	}
}

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

func Debug(msg string, kv ...any) {
	activeSugar().Debugw(msg, kv...)
}

func Info(msg string, kv ...any) {
	activeSugar().Infow(msg, kv...)
}

func Warn(msg string, kv ...any) {
	activeSugar().Warnw(msg, kv...)
}

func Error(msg string, kv ...any) {
	activeSugar().Errorw(msg, kv...)
}

func Sync() error {
	mu.RLock()
	active := logger
	mu.RUnlock()
	if active != nil {
		return normalizeSyncError(active.Sync())
	}
	return normalizeSyncError(fallbackLogger().Sync())
}

func L() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if logger != nil {
		return logger
	}
	return fallbackLogger().Desugar()
}

func S() *zap.SugaredLogger {
	return activeSugar()
}

func SetLevel(level Level) error {
	mu.RLock()
	active := logger
	mu.RUnlock()
	if active == nil {
		return errors.New("log: not initialized")
	}
	zapLevel, err := validateLevel(level)
	if err != nil {
		return err
	}
	levelCtl.SetLevel(zapLevel)
	return nil
}

func normalizeSyncError(err error) error {
	if err == nil {
		return nil
	}
	text := err.Error()
	if text == "sync /dev/stdout: bad file descriptor" || text == "sync /dev/stderr: bad file descriptor" {
		return nil
	}
	return err
}

func resetForTest() {
	mu.Lock()
	defer mu.Unlock()
	logger = nil
	sugar = nil
	inited = false
	levelCtl = zap.NewAtomicLevelAt(zap.InfoLevel)
	fallbackSugar = nil
	fallbackOnce = sync.Once{}
	fallbackOut = os.Stdout
	consoleOut = os.Stdout
}

func setFallbackOutputForTest(w io.Writer) {
	fallbackSugar = nil
	fallbackOnce = sync.Once{}
	fallbackOut = w
}

func setConsoleOutputForTest(w io.Writer) {
	consoleOut = w
}
