package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Conf *Config

// Config 配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

func Init(dir string) (*viper.Viper, error) {
	v := viper.New()

	pflag.StringP("env", "e", "dev", "mode: dev, beta, prod")
	pflag.Parse()

	v.BindPFlags(pflag.CommandLine)

	// 支持环境变量覆盖 前缀 MOKA_
    v.SetEnvPrefix("MOKA")
    v.AutomaticEnv()

	env  := v.GetString("env")
	path := buildPath(dir, env)

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	Conf = &Config{}
	if err := v.Unmarshal(Conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return v, nil
}

func buildPath(dir, env string) string {
	if env == "dev" {
		return filepath.Join(dir, "config.yml")
	}
	return filepath.Join(dir, fmt.Sprintf("config.%s.yml", env))
}
