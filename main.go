package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/xiaowumin-mark/AMLX/app"
	"github.com/xiaowumin-mark/AMLX/config"
	"github.com/xiaowumin-mark/AMLX/logx"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	logger, err := logx.Init(cfg.Log)
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		logger.Error("init app failed", "error", err)
		return
	}

	logger.Info("server starting", "port", cfg.Server.Port)
	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
	}
}
