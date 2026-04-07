package drivenadapters

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/config"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/s3.go -destination ../tests/mock_drivenadapters/s3_mock.go

// S3Object S3对象信息
type S3Object struct {
	Key          string    // 对象键（路径）
	Size         int64     // 文件大小（字节）
	LastModified time.Time // 最后修改时间
	ETag         string    // ETag
}

// S3ObjectMetadata S3对象元数据
type S3ObjectMetadata struct {
	ContentType   string            // 内容类型
	ContentLength int64             // 内容长度
	LastModified  time.Time         // 最后修改时间
	ETag          string            // ETag
	Metadata      map[string]string // 自定义元数据
}

// S3Adapter S3适配器接口
type S3Adapter interface {
	// ValidateBucket 验证Bucket是否存在且可访问
	ValidateBucket(ctx context.Context, bucket string) error

	// ValidatePath 验证路径是否可访问
	ValidatePath(ctx context.Context, bucket, path string) error

	// ListObjects 列出指定路径下的所有对象
	ListObjects(ctx context.Context, bucket, prefix string) ([]S3Object, error)

	// GetObject 获取指定对象的内容
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// GetObjectMetadata 获取对象元数据
	GetObjectMetadata(ctx context.Context, bucket, key string) (*S3ObjectMetadata, error)

	// GeneratePresignedURL 生成预签名下载URL
	GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error)

	// Upload Upload file to S3
	Upload(ctx context.Context, bucket, key string, body io.ReadSeeker) error

	// DeleteObject 删除文件
	DeleteObject(ctx context.Context, bucket, key string) error

	// CopyObject 复制文件
	CopyObject(ctx context.Context, bucket, srcKey, dstKey string) error
}

// s3AdapterImpl S3适配器实现
type s3AdapterImpl struct {
	client *s3.S3
}

// NewS3Adapter 创建S3适配器实例
func NewS3Adapter(cfg *config.S3Config) (S3Adapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("S3 config is nil")
	}

	awsConfig := &aws.Config{
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		DisableSSL:       aws.Bool(!cfg.UseSSL),
		S3ForcePathStyle: aws.Bool(true), // 支持MinIO等S3兼容服务
	}

	if cfg.SkipVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		awsConfig.HTTPClient = &http.Client{Transport: tr}
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &s3AdapterImpl{
		client: s3.New(sess),
	}, nil
}

// ValidateBucket 验证Bucket是否存在且可访问
func (a *s3AdapterImpl) ValidateBucket(ctx context.Context, bucket string) error {
	_, err := a.client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("bucket validation failed: %w", err)
	}
	return nil
}

// ValidatePath 验证路径是否可访问
func (a *s3AdapterImpl) ValidatePath(ctx context.Context, bucket, path string) error {
	// 尝试列出路径下的对象（最多1个）
	_, err := a.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(path),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}
	return nil
}

// ListObjects 列出指定路径下的所有对象
func (a *s3AdapterImpl) ListObjects(ctx context.Context, bucket, prefix string) ([]S3Object, error) {
	var objects []S3Object

	err := a.client.ListObjectsV2PagesWithContext(ctx,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(prefix),
		},
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				// 跳过目录标记（以/结尾且大小为0）
				if obj.Key != nil && obj.Size != nil && *obj.Size == 0 && len(*obj.Key) > 0 && (*obj.Key)[len(*obj.Key)-1] == '/' {
					continue
				}

				objects = append(objects, S3Object{
					Key:          aws.StringValue(obj.Key),
					Size:         aws.Int64Value(obj.Size),
					LastModified: aws.TimeValue(obj.LastModified),
					ETag:         aws.StringValue(obj.ETag),
				})
			}
			return !lastPage
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return objects, nil
}

// GetObject 获取指定对象的内容
func (a *s3AdapterImpl) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	result, err := a.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return result.Body, nil
}

// GetObjectMetadata 获取对象元数据
func (a *s3AdapterImpl) GetObjectMetadata(ctx context.Context, bucket, key string) (*S3ObjectMetadata, error) {
	result, err := a.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	metadata := &S3ObjectMetadata{
		ContentType:   aws.StringValue(result.ContentType),
		ContentLength: aws.Int64Value(result.ContentLength),
		LastModified:  aws.TimeValue(result.LastModified),
		ETag:          aws.StringValue(result.ETag),
		Metadata:      make(map[string]string),
	}

	for k, v := range result.Metadata {
		metadata.Metadata[k] = aws.StringValue(v)
	}

	return metadata, nil
}

// GeneratePresignedURL 生成预签名下载URL
func (a *s3AdapterImpl) GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	req, _ := a.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return urlStr, nil
}

// Upload 上传文件
func (a *s3AdapterImpl) Upload(ctx context.Context, bucket, key string, body io.ReadSeeker) error {
	_, err := a.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// DeleteObject 删除文件
func (a *s3AdapterImpl) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := a.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// CopyObject 复制文件
func (a *s3AdapterImpl) CopyObject(ctx context.Context, bucket, srcKey, dstKey string) error {
	_, err := a.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(bucket + "/" + srcKey),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}
	return nil
}
