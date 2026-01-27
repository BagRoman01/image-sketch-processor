package routers

import (
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(serviceInjector *injectors.ServiceInjector) *gin.Engine {
	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // или конкретные домены
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, MaxAge: 12 * time.Hour,
	}))
	api := r.Group("/api")
	{
		RegisterPingRoutes(api, serviceInjector.FilesSrv)
	}
	return r
}
