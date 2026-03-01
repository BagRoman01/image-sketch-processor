package models

type UploadResponse struct {
	Message    string `json:"message"`
	Key        string `json:"key"`
	URL        string `json:"url"`
	Size       int64  `json:"size"`
	TaskID     string `json:"task_id"`
	TaskStatus string `json:"task_status"`
}

type FileInfo struct {
	FileName string  `json:"file_name"`
	Content  Content `json:"content"`
}

type S3FileInfo struct {
	FileInfo
	FileKey string `json:"file_key"`
	FileID  string `json:"file_id"`
}
type Content struct {
	ContentLength int64  `json:"content_size"`
	ContentType   string `json:"content_type"`
}
