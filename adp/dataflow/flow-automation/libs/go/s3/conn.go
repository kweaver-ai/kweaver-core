package s3

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
)

type S3Connection struct {
	client          *s3.S3
	Name            string `json:"name"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	BucketName      string `json:"bucketName"`
	SkipVerify      bool   `json:"skipVerify"`
	DisableHTTPS    bool   `json:"disableHTTPS"`
}

func (c *S3Connection) InitClient() error {
	if c.client != nil {
		return nil
	}

	awsConfig := &aws.Config{
		Endpoint:         aws.String(c.Endpoint),
		Region:           aws.String(c.Region),
		Credentials:      credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		DisableSSL:       aws.Bool(c.DisableHTTPS),
		S3ForcePathStyle: aws.Bool(true),
	}

	if c.SkipVerify && !c.DisableHTTPS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		awsConfig.HTTPClient = &http.Client{Transport: tr}
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	c.client = s3.New(sess)

	return nil
}

// GetDownloadURL 生成一个限时的预签名下载链接
func (c *S3Connection) GetDownloadURL(ctx context.Context, key string, expires int64) (string, error) {
	if err := c.InitClient(); err != nil {
		return "", err
	}

	req, _ := c.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	if expires <= 0 {
		expires = 3600
	}

	duration := time.Duration(expires) * time.Second
	urlStr, err := req.Presign(duration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return urlStr, nil
}

func (c *S3Connection) DeleteObject(ctx context.Context, key string) error {
	if c.client == nil {
		return fmt.Errorf("s3 client is not initialized")
	}

	_, err := c.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	return err
}

func (c *S3Connection) UploadObject(ctx context.Context, key string, file io.Reader, size int64) error {
	if c.client == nil {
		return fmt.Errorf("s3 client is not initialized")
	}

	_, err := c.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.BucketName),
		Key:           aws.String(key),
		Body:          aws.ReadSeekCloser(file),
		ContentLength: aws.Int64(size),
	})

	return err
}

func (c *S3Connection) GetObjectSize(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("s3 client is not initialized")
	}

	resp, err := c.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to head object: %w", err)
	}

	return *resp.ContentLength, nil
}

func (c *S3Connection) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if c.client == nil {
		return nil, fmt.Errorf("s3 client is not initialized")
	}

	resp, err := c.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return resp.Body, nil
}

// GetUploadURL 生成一个限时的预签名上传链接
func (c *S3Connection) GetUploadURL(ctx context.Context, key string, expires int64) (url string, err error) {
	if err := c.InitClient(); err != nil {
		return "", err
	}

	req, _ := c.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	if expires <= 0 {
		expires = 3600
	}

	duration := time.Duration(expires) * time.Second
	urlStr, err := req.Presign(duration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return urlStr, nil
}
