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
