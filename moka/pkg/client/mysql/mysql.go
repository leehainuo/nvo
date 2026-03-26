package mysql

import (
	"fmt"
	"moka/pkg/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var client *gorm.DB

type Config struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

func Init() error {
	var c Config
	if err := config.Viper.UnmarshalKey("mysql", &c); err != nil {
		return fmt.Errorf("failed to unmarshal mysql config: %w", err)
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
		return fmt.Errorf("failed to connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get mysql instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}

	client = db

	return nil
}

func Close() error {
	if client == nil {
		return nil
	}

	sqlDB, err := client.DB()
	if err != nil {
		return fmt.Errorf("failed to get mysql instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close mysql: %w", err)
	}

	return nil
}

func Client() *gorm.DB {
	return client
}