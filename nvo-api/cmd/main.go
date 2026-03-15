package main

import (
	"fmt"

	"nvo-api/core"
	"nvo-api/core/config"
	"nvo-api/core/log"
	"nvo-api/internal/system"
)

func main() {
	configPath := "config.yml"

	// 1. 加载配置
	c, _, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化全局 logger
	if err := log.Init(c.Log); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer log.Sync() // 应用关闭时同步日志

	// 3. 初始化 Pocket
	pocket := core.NewPocketBuilder(configPath).
		WithEnforcer(). // 启用 Casbin 权限控制
		MustBuild()
	defer pocket.Close()

	log.Info("Application initialized successfully")

	api := pocket.GinEngine.Group("/api/v1")
	system.RegisterModules(api, pocket)

	// 启动服务
	addr := fmt.Sprintf("%s:%d", pocket.Config.Server.Host, pocket.Config.Server.Port)
	log.Info(fmt.Sprintf("Server starting on %s", addr))

	if err := pocket.GinEngine.Run(addr); err != nil {
		log.Fatal(fmt.Sprintf("Failed to start server: %v", err))
	}
}
