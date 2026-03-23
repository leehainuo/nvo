package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config Core 配置结构
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Auth   AuthConfig   `mapstructure:"auth"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

type LogConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Format     string `mapstructure:"format"`      // json, console
	OutputPath string `mapstructure:"output_path"` // stdout, 或文件路径
}

type AuthConfig struct {
	ModelPath  string   `mapstructure:"model_path"`
	PolicyPath string   `mapstructure:"policy_path"`
	Whitelist  []string `mapstructure:"whitelist"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, *viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 设置默认值
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 返回 viper 实例，让各客户端包自己读取配置
	return &c, v, nil
}

// setDefaults 设置默认配置（仅核心配置）
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.name", "nvo-api")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output_path", "stdout")
}
