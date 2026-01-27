package services

import (
	"github.com/BagRoman01/image-sketch-processor/internal/models"
)

type FilesService struct {
}

func NewFilesService() *FilesService {
	return &FilesService{}
}

func (f *FilesService) Pong() *models.PongResponse {
	return &models.PongResponse{
		Msg: "pong",
	}
}
