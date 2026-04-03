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
