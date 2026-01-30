package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/BagRoman01/image-sketch-processor/docs"
	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/routers"
)

// @title           Image Sketch Processor API
// @version         1.0
// @description     API для обработки изображений
// @host            localhost:8000
// @BasePath        /api/
func main() {
	cfg := config.NewConfig()
	mainCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	logger, err := logging.InitLogger(&cfg.LogConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("initializing service dependencies")
	serviceInjector, err := injectors.NewServiceInjector(mainCtx, cfg)
	if err != nil {
		logger.Error("failed to initialize services", "error", err)
		os.Exit(1)
	}
	logger.Info("service dependencies initialized successfully")

	r := routers.SetupRouter(serviceInjector)

	addr := fmt.Sprintf(
		"%s:%d",
		cfg.InstanceConfig.Host,
		cfg.InstanceConfig.Port,
	)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("server starting",
		"address", addr,
		"host", cfg.InstanceConfig.Host,
		"port", cfg.InstanceConfig.Port,
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("waiting for shutdown signal...")
	<-mainCtx.Done()

	logger.Info("starting graceful shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "error", err)
	} else {
		logger.Info("graceful shutdown completed")
	}

}
