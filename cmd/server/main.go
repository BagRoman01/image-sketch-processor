package main

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/BagRoman01/image-sketch-processor/docs"
	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/routers"
)

// @title           Image Sketch Processor API
// @version         1.0
// @description     API для обработки изображений
// @host            localhost:8000
// @BasePath        /api/
func main() {
	cfg := config.NewConfig()
	ctx := context.Background()
	serviceInjector, err := injectors.NewServiceInjector(ctx, cfg)
	if err != nil {
		panic(err)
	}
	r := routers.SetupRouter(serviceInjector)

	addr := fmt.Sprintf(
		"%s:%d",
		cfg.InstanceConfig.Host,
		cfg.InstanceConfig.Port,
	)

	fmt.Println(addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// appLogger.WithData(map[string]interface{}{
		// 	"error": err.Error(),
		// }).Fatal("Ошибка запуска сервера")
	}
}
