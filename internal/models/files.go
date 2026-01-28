package models

type UploadResponse struct {
	Message  string `json:"message"`
	Key      string `json:"key"`
	Location string `json:"location"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
}

type FileMetadata struct {
	ContentType   string
	ContentLength int64
	FileName      string
}
