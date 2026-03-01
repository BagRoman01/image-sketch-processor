package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

	// Инициализация логгера
	_, err := logging.InitLogger(&cfg.LogConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Контекст для всего приложения (отменяется по сигналу)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Контекст для инициализации с таймаутом
	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	defer initCancel()

	slog.Info("initializing service dependencies")
	serviceInjector, err := injectors.NewServiceInjector(initCtx, cfg)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			slog.Error("service initialization timed out after 30s")
		case errors.Is(err, context.Canceled):
			slog.Error("service initialization cancelled by signal")
		default:
			slog.Error("failed to initialize services", "error", err)
		}
		os.Exit(1)
	}
	slog.Info("service dependencies initialized successfully")

	// Роутер
	r := routers.SetupRouter(serviceInjector)

	// HTTP сервер
	srv := &http.Server{
		Addr:    cfg.InstanceConfig.Address(),
		Handler: r,
	}

	// Канал для ошибок сервера
	serverErrors := make(chan error, 1)

	// Запуск сервера в горутине
	go func() {
		slog.Info("server starting",
			"address", cfg.InstanceConfig.Address(),
			"host", cfg.InstanceConfig.Host,
			"port", cfg.InstanceConfig.Port,
		)
		serverErrors <- srv.ListenAndServe()
	}()

	// Ожидание сигнала завершения или ошибки сервера
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 1*time.Minute,
	)
	defer shutdownCancel()

	// Останавливаем HTTP сервер
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown failed", "error", err)
	} else {
		slog.Info("HTTP server shutdown completed")
	}

	if err := serviceInjector.Shutdown(shutdownCtx); err != nil {
		slog.Error("service injector shutdown failed", "error", err)
		// Проверяем, не превысили ли таймаут
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Warn(
				"shutdown timed out, some resources may not be closed properly",
			)
		}
	} else {
		slog.Info("service injector shutdown completed")
	}
}
