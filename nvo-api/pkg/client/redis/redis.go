package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// InitRedis 初始化Redis连接
func InitRedis(c Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
		PoolSize: c.PoolSize,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}

// InitFromViper 从 Viper 配置中初始化 Redis
// 客户端包自己负责读取配置，实现完全解耦
func InitFromViper(v *viper.Viper, key string) (*redis.Client, error) {
	var c Config
	if err := v.UnmarshalKey(key, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}
	return InitRedis(c)
}
