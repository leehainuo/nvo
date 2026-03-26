package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OptionFunc func (*Config)

func New(host string, port int, opts ...OptionFunc) (*redis.Client, error) {
	c := &Config{}

	for _, opt := range opts {
		opt(c)
	} 

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
		PoolSize: c.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}

func WithPassword(password string) OptionFunc {
	return func(c *Config) {
		c.Password = password
	}
}

func WithDB(db int) OptionFunc {
	return func(c *Config) {
		c.DB = db
	}
}

func WithPoolSize(size int) OptionFunc {
	return func(c *Config) {
		c.PoolSize = size
	}
}