package routers

import (
	"log/slog"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(serviceInjector *injectors.ServiceInjector) *gin.Engine {
	slog.Debug("setting up router and middlewares")

	r := gin.New()
	r.Use(middlewares.LoggingMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, MaxAge: 12 * time.Hour,
	}))

	slog.Debug("registering API routes")

	api := r.Group("/api")
	{
		RegisterFilesRoutes(api, serviceInjector)
		RegisterTasksRoutes(api, serviceInjector)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	slog.Info("router configured successfully")

	return r
}
