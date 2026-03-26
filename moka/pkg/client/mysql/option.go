package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type OptionFunc func (*Config)

func New(host, username, database string, opts ...OptionFunc) (*gorm.DB, error) {
	c := &Config{}

	for _, opt := range opts {
		opt(c)
	}
 
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        c.Username,
        c.Password,
        c.Host,
        c.Port,
        c.Database,
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect mysql: %w", err)
    }
 
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get mysql instance: %w", err)
    }

    if c.MaxIdleConns > 0 {
        sqlDB.SetMaxIdleConns(c.MaxIdleConns)
    }
    if c.MaxOpenConns > 0 {
        sqlDB.SetMaxOpenConns(c.MaxOpenConns)
    }
    if c.ConnMaxLifetime > 0 {
        sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)
    }

	if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping mysql: %w", err)
    }

	return db, nil
}

func WithPassword(password string) OptionFunc {
	return func(c *Config) {
		c.Password = password
	}
}

func WithDatabase(database string) OptionFunc {
	return func(c *Config) {
		c.Database = database
	}
}

func WithMaxIdleConns(n int) OptionFunc {
	return func(c *Config) {
		c.MaxIdleConns = n
	}
}

func WithMaxOpenConns(n int) OptionFunc {
	return func(c *Config) {
		c.MaxOpenConns = n
	}
}

func WithConnMaxLifetime(seconds int) OptionFunc {
    return func(c *Config) {
        c.ConnMaxLifetime = seconds
    }
}
