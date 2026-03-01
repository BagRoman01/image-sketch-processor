package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
)

func main() {
	cfg := config.NewConfig()

	_, err := logging.InitLogger(&cfg.LogConfig)
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	appCtx, appCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer appCancel()

	initCtx, initCancel := context.WithTimeout(appCtx, 1*time.Minute)
	defer initCancel()

	slog.Info("initializing service dependencies")
	serviceInjector, err := injectors.NewServiceInjector(initCtx, cfg)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			slog.Error("service initialization timed out after 1 minute")
		case errors.Is(err, context.Canceled):
			slog.Error("service initialization cancelled by signal")
		default:
			slog.Error("failed to initialize services", "error", err)
		}
		os.Exit(1)
	}
	slog.Info("service dependencies initialized successfully")

	if serviceInjector.ProcessingService != nil {
		go func() {
			slog.Info("starting processor")
			if err := serviceInjector.ProcessingService.Start(appCtx); err != nil {
				slog.Error("processor failed", "error", err)
				appCancel()
			}
		}()
	}

	<-appCtx.Done()
	slog.Info("shutdown signal received, shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 1*time.Minute,
	)
	defer shutdownCancel()

	if err := serviceInjector.Shutdown(shutdownCtx); err != nil {
		slog.Error("service injector shutdown failed", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Warn(
				"shutdown timed out, some resources may not be closed properly",
			)
		}
	} else {
		slog.Info("service injector shutdown completed")
	}

	slog.Info("application stopped")
}
