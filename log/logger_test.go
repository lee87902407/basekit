package log

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInfoUsesFallbackConsoleLoggerBeforeInit(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
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

func TestInitRejectsSecondInitialization(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
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
	t.Cleanup(resetForTest)
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
}

func TestOutputModeBothWritesConsoleAndFile(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
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
}

func TestSetLevelChangesActiveLevel(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
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

func TestSyncWorksWithFallbackLoggerBeforeInit(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
	if err := Sync(); err != nil {
		t.Fatalf("Sync error = %v", err)
	}
}

func TestInitRejectsInvalidLevel(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
	cfg := DefaultConfig()
	cfg.Level = Level(255)

	if err := Init(cfg); err == nil {
		t.Fatal("Init should reject invalid level")
	}
}

func TestSetLevelRejectsInvalidLevel(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
	console := &bytes.Buffer{}
	setConsoleOutputForTest(console)

	if err := Init(DefaultConfig()); err != nil {
		t.Fatalf("Init error = %v", err)
	}
	if err := SetLevel(Level(255)); err == nil {
		t.Fatal("SetLevel should reject invalid level")
	}
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	text := string(data)
	if !strings.Contains(text, want) {
		t.Fatalf("file %q content = %q, want contains %q", path, text, want)
	}
}
