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
