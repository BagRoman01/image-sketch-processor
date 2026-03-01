package models

import "time"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID          string     `json:"id"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt time.Time  `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type S3FileTask struct {
	Task
	ProcessedKey string     `json:"processed_key,omitempty"`
	DownloadURL  string     `json:"download_url,omitempty"`
	S3FileInfo   S3FileInfo `json:"file_info"`
}
