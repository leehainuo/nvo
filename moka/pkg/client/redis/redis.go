package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var Client *redis.Client

// Config Redis配置
type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init(v *viper.Viper, key string) error {
	var c Config
	if err := v.UnmarshalKey(key, &c); err != nil {
		return fmt.Errorf("failed to unmarshal redis config: %w", err)
	}
		Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
		PoolSize: c.PoolSize,
	})

	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

func Close() error {
	if Client == nil {
		return nil
	}

	if err := Client.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %v\n", err)
	}

	return nil
}