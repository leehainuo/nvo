package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunConfig struct {
	Bucket          string `mapstructure:"bucket"`
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
}

type aliyunStorage struct {
	bucket *oss.Bucket
	client *oss.Client
}

func initAliyun(c *Config) (Storager, error) {
	if c.Aliyun.Bucket == "" {
		return nil, fmt.Errorf("aliyun bucket is required")
	}
	if c.Aliyun.Endpoint == "" {
		return nil, fmt.Errorf("aliyun endpoint is required")
	}
	if c.Aliyun.AccessKeyID == "" {
		return nil, fmt.Errorf("aliyun access_key_id is required")
	}
	if c.Aliyun.AccessKeySecret == "" {
		return nil, fmt.Errorf("aliyun access_key_secret is required")
	}

	client, err := oss.New(
		c.Aliyun.Endpoint,
		c.Aliyun.AccessKeyID,
		c.Aliyun.AccessKeySecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create aliyun oss client: %w", err)
	}

	bucket, err := client.Bucket(c.Aliyun.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	aliyun := &aliyunStorage{
		bucket: bucket,
		client: client,
	}

	return aliyun, nil
}

func (a *aliyunStorage) Upload(c context.Context, key string, reader io.Reader) (string, error) {
	if err := a.bucket.PutObject(key, reader); err != nil {
		return "", fmt.Errorf("failed to upload to aliyun oss: %w", err)
	}

	return a.URL(c, key)
}

func (a *aliyunStorage) Download(c context.Context, key string) (io.ReadCloser, error) {
	body, err := a.bucket.GetObject(key)
	if err != nil {
		return nil, fmt.Errorf("failed to download from aliyun oss: %w", err)
	}

	return body, nil
}

func (a *aliyunStorage) Exists(c context.Context, key string) (bool, error) {
	exists, err := a.bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return exists, nil
}

func (a *aliyunStorage) URL(c context.Context, key string) (string, error) {
	url := fmt.Sprintf("https://%s.%s/%s",
		a.bucket.BucketName,
		a.client.Config.Endpoint,
		key,
	)
	return url, nil
}

func (a *aliyunStorage) Delete(c context.Context, key string) error {
	if err := a.bucket.DeleteObject(key); err != nil {
		return fmt.Errorf("failed to delete from aliyun oss: %w", err)
	}

	return nil
}

func (a *aliyunStorage) Close() error {
	return nil
}