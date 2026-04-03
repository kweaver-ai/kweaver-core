package adapter

import "context"

type OSSAdapter interface {
	GetVendorType() VendorType

	CheckConnection(ctx context.Context) error

	GetHeadURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error)

	GetUploadURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error)

	GetPutUploadURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error)

	InitMultipartUpload(ctx context.Context, objectKey string, fileSize int64) (*MultipartInitResult, error)

	GetUploadPartURL(ctx context.Context, objectKey, uploadID string, partNumber int, validSeconds int64) (*PresignedURL, error)

	GetCompleteMultipartUpload(ctx context.Context, objectKey, uploadID string, parts []PartInfo) (*PresignedURL, error)

	AbortMultipartUpload(ctx context.Context, objectKey, uploadID string) error

	GetDownloadURL(ctx context.Context, objectKey string, validSeconds int64, fileName string) (*PresignedURL, error)

	GetDeleteURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error)
}
