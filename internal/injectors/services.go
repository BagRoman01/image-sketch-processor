package injectors

import (
	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
)

type ServiceInjector struct {
	FilesSrv *services.FilesService
}

func NewServiceInjector(cfg *config.ServiceConfig) *ServiceInjector {
	filesSrv := services.NewFilesService()
	return &ServiceInjector{FilesSrv: filesSrv}
}
