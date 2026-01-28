// internal/services/files_service.go
package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/oklog/ulid/v2"
)

type S3storageService struct {
	S3Repo  *repositories.S3Repository
	entropy *ulid.LockedMonotonicReader
}

func NewS3storageService(s3Repo *repositories.S3Repository) *S3storageService {
	return &S3storageService{
		S3Repo: s3Repo,
		entropy: &ulid.LockedMonotonicReader{
			MonotonicReader: ulid.Monotonic(rand.Reader, 0),
		},
	}
}

func (s *S3storageService) UploadFileStream(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
) (*manager.UploadOutput, string, error) {
	ext := filepath.Ext(filepath.Base(fileHeader.Filename))
	ulidID := ulid.MustNew(ulid.Timestamp(time.Now()), s.entropy).String()
	filename := fmt.Sprintf("%s%s", ulidID, ext)
	key := fmt.Sprintf("uploads/%s", filename)

	result, err := s.S3Repo.UploadFileStream(ctx, fileHeader, key)
	return result, key, err
}
