package adapter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MinPartSize     = 5 * 1024 * 1024
	MaxPartSize     = 5 * 1024 * 1024 * 1024
	MaxParts        = 10000
	DefaultPartSize = 64 * 1024 * 1024
)

type MinIOAdapter struct {
	client     *minio.Client
	core       *minio.Core
	config     StorageConfig
	bucketName string
}

func NewMinIOAdapter(config StorageConfig) (*MinIOAdapter, error) {
	endpoint := config.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	options := &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.AccessKeySecret, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	}

	// 火山云 TOS 必须使用 VirtualHostStyle（DNS 方式），不支持 PathStyle
	// https://{bucketname}.tos-s3-cn-beijing.volces.com
	if config.VendorType == VendorTOS {
		options.BucketLookup = minio.BucketLookupDNS
	}

	// 对于 ECEPH 存储，如果使用 HTTPS，跳过证书验证
	// 因为私有化部署可能使用自签名证书或没有购买证书
	if config.VendorType == VendorECEPH && config.UseSSL {
		options.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client, err := minio.New(endpoint, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	core, err := minio.NewCore(endpoint, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio core client: %w", err)
	}

	return &MinIOAdapter{
		client:     client,
		core:       core,
		config:     config,
		bucketName: config.BucketName,
	}, nil
}

func (a *MinIOAdapter) GetVendorType() VendorType {
	return a.config.VendorType
}

func (a *MinIOAdapter) CheckConnection(ctx context.Context) error {
	exists, err := a.client.BucketExists(ctx, a.bucketName)
	if err != nil {
		return fmt.Errorf("connection check failed: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", a.bucketName)
	}
	return nil
}

func (a *MinIOAdapter) GetHeadURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error) {
	expires := time.Duration(validSeconds) * time.Second

	presignedURL, err := a.client.PresignedHeadObject(ctx, a.bucketName, objectKey, expires, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate head url: %w", err)
	}

	return &PresignedURL{
		Method:  http.MethodHead,
		URL:     presignedURL.String(),
		Headers: map[string]string{},
	}, nil
}

func (a *MinIOAdapter) GetUploadURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error) {
	policy := minio.NewPostPolicy()
	policy.SetBucket(a.bucketName)
	policy.SetKey(objectKey)
	policy.SetExpires(time.Now().UTC().Add(time.Duration(validSeconds) * time.Second))

	presignedURL, formData, err := a.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate post upload url: %w", err)
	}

	return &PresignedURL{
		Method:    http.MethodPost,
		URL:       presignedURL.String(),
		Headers:   map[string]string{},
		FormField: formData,
	}, nil
}

func (a *MinIOAdapter) GetPutUploadURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error) {
	expires := time.Duration(validSeconds) * time.Second

	presignedURL, err := a.client.PresignedPutObject(ctx, a.bucketName, objectKey, expires)
	if err != nil {
		return nil, fmt.Errorf("failed to generate put upload url: %w", err)
	}

	return &PresignedURL{
		Method:  http.MethodPut,
		URL:     presignedURL.String(),
		Headers: map[string]string{},
	}, nil
}

func (a *MinIOAdapter) InitMultipartUpload(ctx context.Context, objectKey string, fileSize int64) (*MultipartInitResult, error) {
	partSize := a.calculatePartSize(fileSize)

	uploadID, err := a.core.NewMultipartUpload(ctx, a.bucketName, objectKey, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init multipart upload: %w", err)
	}

	return &MultipartInitResult{
		UploadID: uploadID,
		PartSize: partSize,
	}, nil
}

func (a *MinIOAdapter) GetUploadPartURL(ctx context.Context, objectKey, uploadID string, partNumber int, validSeconds int64) (*PresignedURL, error) {
	expires := time.Duration(validSeconds) * time.Second

	presignedURL, err := a.client.Presign(ctx, http.MethodPut, a.bucketName, objectKey, expires, url.Values{
		"partNumber": []string{fmt.Sprintf("%d", partNumber)},
		"uploadId":   []string{uploadID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate part upload url: %w", err)
	}

	return &PresignedURL{
		Method:  http.MethodPut,
		URL:     presignedURL.String(),
		Headers: map[string]string{},
	}, nil
}

func (a *MinIOAdapter) GetCompleteMultipartUpload(ctx context.Context, objectKey, uploadID string, parts []PartInfo) (*PresignedURL, error) {
	var xmlParts strings.Builder
	xmlParts.WriteString("<CompleteMultipartUpload>")
	for _, p := range parts {
		etag := p.ETag
		if !strings.HasPrefix(etag, "\"") {
			etag = fmt.Sprintf("\"%s\"", etag)
		}
		xmlParts.WriteString(fmt.Sprintf("<Part><PartNumber>%d</PartNumber><ETag>%s</ETag></Part>",
			p.PartNumber, etag))
	}
	xmlParts.WriteString("</CompleteMultipartUpload>")

	expires := 1 * time.Hour
	presignedURL, err := a.client.Presign(ctx, http.MethodPost, a.bucketName, objectKey, expires, url.Values{
		"uploadId": []string{uploadID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate complete upload url: %w", err)
	}

	return &PresignedURL{
		Method: http.MethodPost,
		URL:    presignedURL.String(),
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
		Body: xmlParts.String(),
	}, nil
}

func (a *MinIOAdapter) AbortMultipartUpload(ctx context.Context, objectKey, uploadID string) error {
	err := a.core.AbortMultipartUpload(ctx, a.bucketName, objectKey, uploadID)
	if err != nil {
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}
	return nil
}

func (a *MinIOAdapter) GetDownloadURL(ctx context.Context, objectKey string, validSeconds int64, fileName string) (*PresignedURL, error) {
	expires := time.Duration(validSeconds) * time.Second

	reqParams := make(url.Values)
	if fileName != "" {
		disposition := fmt.Sprintf("attachment; filename*=utf-8''%s", url.PathEscape(fileName))
		reqParams.Set("response-content-disposition", disposition)
	}

	presignedURL, err := a.client.PresignedGetObject(ctx, a.bucketName, objectKey, expires, reqParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download url: %w", err)
	}

	return &PresignedURL{
		Method:  http.MethodGet,
		URL:     presignedURL.String(),
		Headers: map[string]string{},
	}, nil
}

func (a *MinIOAdapter) GetDeleteURL(ctx context.Context, objectKey string, validSeconds int64) (*PresignedURL, error) {
	expires := time.Duration(validSeconds) * time.Second

	presignedURL, err := a.client.Presign(ctx, http.MethodDelete, a.bucketName, objectKey, expires, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate delete url: %w", err)
	}

	return &PresignedURL{
		Method:  http.MethodDelete,
		URL:     presignedURL.String(),
		Headers: map[string]string{},
	}, nil
}

func (a *MinIOAdapter) calculatePartSize(fileSize int64) int64 {
	if fileSize <= 0 {
		return DefaultPartSize
	}

	partSize := fileSize / MaxParts
	if partSize < MinPartSize {
		partSize = MinPartSize
	}
	if partSize > MaxPartSize {
		partSize = MaxPartSize
	}

	const MB = 1024 * 1024
	partSize = ((partSize + MB - 1) / MB) * MB

	return partSize
}
