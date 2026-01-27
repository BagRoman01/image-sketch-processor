package main

import (
	"fmt"
	"net/http"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/routers"
)

func main() {
	cfg := config.NewServiceConfig()
	serviceInjector := injectors.NewServiceInjector(cfg)
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
