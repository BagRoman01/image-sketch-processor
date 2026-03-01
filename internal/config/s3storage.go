package config

type S3StorageConfig struct {
	Region            string `yaml:"region" envconfig:"s3_region"`
	Bucket            string `yaml:"bucket" envconfig:"s3_bucket"`
	AccessKeyID       string `yaml:"access_key_id" envconfig:"s3_access_key_id"`
	SecretAccessKey   string `yaml:"secret_access_key" envconfig:"s3_secret_access_key"`
	Endpoint          string `yaml:"endpoint" envconfig:"s3_endpoint"`
	UseSSL            bool   `yaml:"use_ssl" envconfig:"s3_use_ssl"`
	MaxUploadSize     int64  `yaml:"max_upload_size" envconfig:"s3_max_upload_size"`
	ChunkUploadSize   int64  `yaml:"chunk_upload_size" envconfig:"s3_chunk_upload_size"`
	UploadConcurrency uint16 `yaml:"upload_concurrency" envconfig:"upload_concurrency"`
}

func NewS3StorageConfig() *S3StorageConfig {
	return &S3StorageConfig{
		Region:            "ru-central1",
		Bucket:            "files",
		AccessKeyID:       "s3_access_key_id",
		SecretAccessKey:   "s3_secret_access_key",
		Endpoint:          "http://s3-storage:9000",
		UseSSL:            false,
		MaxUploadSize:     104857600, // 100 MB
		ChunkUploadSize:   5242880,   // 5 MB
		UploadConcurrency: 5,
	}
}
