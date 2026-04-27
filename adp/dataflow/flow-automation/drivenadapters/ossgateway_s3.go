package drivenadapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/s3"
)

type ossGatetWayS3 struct {
	s3 *s3.S3
}

func NewOssGatewayS3() OssGateWay {
	return &ossGatetWayS3{s3: s3.NewS3()}
}

// DeleteFile implements OssGateWay.
func (o *ossGatetWayS3) DeleteFile(ctx context.Context, ossID string, key string, internalRequest bool) error {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return fmt.Errorf("s3 connection not found")
	}

	if err := conn.DeleteObject(ctx, key); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// DownloadFile implements OssGateWay.
func (o *ossGatetWayS3) DownloadFile(ctx context.Context, ossID string, key string, internalRequest bool, opts ...OssOpt) ([]byte, error) {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return nil, fmt.Errorf("s3 connection not found")
	}

	body, err := conn.GetObject(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer body.Close()

	// 读取 body 内容
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}

// DownloadFile2Local implements OssGateWay.
func (o *ossGatetWayS3) DownloadFile2Local(ctx context.Context, ossID string, key string, internalRequest bool, filePath string, opts ...OssOpt) (int64, error) {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return 0, fmt.Errorf("s3 connection not found")
	}

	body, err := conn.GetObject(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get object: %w", err)
	}
	defer body.Close()

	// 读取 body 内容
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, fmt.Errorf("failed to read object body: %w", err)
	}

	// 写入本地文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write file: %w", err)
	}

	return int64(len(data)), nil
}

// GetAvaildOSS implements OssGateWay.
func (o *ossGatetWayS3) GetAvaildOSS(ctx context.Context) (string, error) {
	conn := o.s3.GetAvailableConnection()
	if conn == nil {
		return "", fmt.Errorf("no available s3 connection")
	}

	return conn.Name, nil
}

// GetDownloadURL implements OssGateWay.
func (o *ossGatetWayS3) GetDownloadURL(ctx context.Context, ossID string, key string, expires int64, internalRequest bool, opts ...OssOpt) (string, error) {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return "", fmt.Errorf("s3 connection not found")
	}

	url, err := conn.GetDownloadURL(ctx, key, expires)
	if err != nil {
		return "", fmt.Errorf("failed to get download url: %w", err)
	}

	return url, nil
}

// GetObjectMeta implements OssGateWay.
func (o *ossGatetWayS3) GetObjectMeta(ctx context.Context, ossID string, key string, internalRequest bool, opts ...OssOpt) (int64, error) {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return 0, fmt.Errorf("s3 connection not found")
	}

	size, err := conn.GetObjectSize(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get object size: %w", err)
	}

	return size, nil
}

// NewReader implements OssGateWay.
func (o *ossGatetWayS3) NewReader(ossID string, ossKey string, opts ...OssOpt) *Reader {
	return &Reader{
		og:     o,
		ossID:  ossID,
		ossKey: ossKey,
		opts:   opts,
		cache:  make(map[string]string),
	}
}

// SimpleUpload implements OssGateWay.
func (o *ossGatetWayS3) SimpleUpload(ctx context.Context, ossID string, key string, internalRequest bool, file io.Reader) error {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return fmt.Errorf("s3 connection not found")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := conn.UploadObject(ctx, key, bytes.NewReader(data), int64(len(data))); err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// UploadFile implements OssGateWay.
func (o *ossGatetWayS3) UploadFile(ctx context.Context, ossID string, key string, internalRequest bool, file io.Reader, size int64) error {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return fmt.Errorf("s3 connection not found")
	}

	if err := conn.UploadObject(ctx, key, file, size); err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// GetUploadReq implements OssGateWay.
func (o *ossGatetWayS3) GetUploadReq(ctx context.Context, ossID string, key string, expires int64, internalRequest bool) (*UploadRequest, error) {
	conn := o.s3.GetConnection(ossID)
	if conn == nil {
		return nil, fmt.Errorf("s3 connection not found")
	}

	url, err := conn.GetUploadURL(ctx, key, expires)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload url: %w", err)
	}

	return &UploadRequest{
		Method:  "PUT",
		URL:     url,
		Headers: map[string]string{},
		Expires: expires,
	}, nil
}
