package adapter

type VendorType string

const (
	VendorOSS   VendorType = "OSS"
	VendorOBS   VendorType = "OBS"
	VendorECEPH VendorType = "ECEPH"
	VendorTOS   VendorType = "TOS"
)

type StorageConfig struct {
	StorageID        string
	VendorType       VendorType
	Endpoint         string
	BucketName       string
	AccessKeyID      string
	AccessKeySecret  string
	Region           string
	UseSSL           bool
	InternalEndpoint string
}

type PresignedURL struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	FormField map[string]string `json:"form_field"`
	Body      string            `json:"body"`
}

type MultipartInitResult struct {
	UploadID string `json:"upload_id"`
	PartSize int64  `json:"part_size"`
}

type PartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}
