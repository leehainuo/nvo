package redis

import (
	"context"
	"fmt"
	"moka/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() error {
	var c Config
	if err := config.Viper.UnmarshalKey("redis", &c); err != nil {
		return fmt.Errorf("failed to unmarshal redis config: %w", err)
	}
		client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
		PoolSize: c.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

func Close() error {
	if client == nil {
		return nil
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}

	return nil
}

func Client() *redis.Client {
	return client
}