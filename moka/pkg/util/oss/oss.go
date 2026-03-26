package oss

import (
	"context"
	"fmt"
	"io"
	"moka/pkg/config"
	"sync"
)

type Storager interface {
	// Upload 上传文件
	Upload(c context.Context, key string, reader io.Reader) (string, error)

	// Download 下载文件
	Download(c context.Context, key string) (io.ReadCloser, error)

	// Exists 判断文件是否存在
	Exists(c context.Context, key string) (bool, error)

	// URL 获取文件访问地址
	URL(c context.Context, key string) (string, error)

	// Delete 删除文件
	Delete(c context.Context, key string) error

	// Close 关闭存储客户端，释放资源
	Close() error
}

type Config struct {
    Type   string       `mapstructure:"type"`
    Local  LocalConfig  `mapstructure:"local"`
    Aliyun AliyunConfig `mapstructure:"aliyun"`
}

var (
	err    error
	once   sync.Once
	client Storager
)

func Client() (Storager, error) {
	once.Do(func() {
		var c Config
		if e := config.Viper.UnmarshalKey("oss", &c); e != nil {
			err = fmt.Errorf("failed to unmarshal oss config: %w", e)
			return
		}

		switch c.Type {
		case "local":
			client, err = initLocal(&c)
		case "aliyun":
			client, err = initAliyun(&c)
		default:
			err = fmt.Errorf("unsupported storage type: %s", c.Type)
		}
	})

	return client, err
}

func Close() error {
	if client == nil {
		return nil
	}

	return client.Close()
}