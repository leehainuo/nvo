package main

import (
	"context"
	"fmt"
	"moka/internal/admin/router"
	"moka/pkg/auth/casbin"
	"moka/pkg/client/mysql"
	"moka/pkg/client/redis"
	"moka/pkg/config"
	"moka/pkg/util/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 初始化配置
	viper, err := config.Init("config/admin")
	if err != nil {
		panic(fmt.Sprintf("\033[1;31mFailed to load config: %v\033[0m", err))
	}

	// 初始化日志
	if err := log.Init(viper, "log"); err != nil {
		panic(fmt.Sprintf("\033[1;31mFailed to initialize log: %v\033[0m", err))
	}
	defer log.Sync()

	// 初始化MySQL
	if err := mysql.Init(viper, "mysql"); err != nil {
		panic(fmt.Sprintf("\033[1;31mFailed to initialize mysql: %v\033[0m", err))
	}
	defer mysql.Close()

	// 初始化Redis
	if err := redis.Init(viper, "redis"); err != nil {
		panic(fmt.Sprintf("\033[1;31mFailed to initialize redis: %v\033[0m", err))
	}
	defer redis.Close()

	// 初始化Casbin
	if err := casbin.Init(); err != nil {
		panic(fmt.Sprintf("\033[1;31mFailed to initialize casbin: %v\033[0m", err))
	}
	defer casbin.Close()

	// 初始化Gin路由
	router := router.Init()

	// 创建HTTP Server
	addr   := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 优雅启动
	go func() {
		log.Info("Moka moka ~")
		log.Info("Starting server ...")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println()
	log.Info("Shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("bye~")
}
