package main

import (
	"context"
	"os"
	"os/signal"
	"pixelpunk/internal/bootstrap"
	"pixelpunk/pkg/logger"
	"syscall"
	"time"
)

// Version 应用版本号，可通过 ldflags 在编译时注入
// 构建命令示例: go build -ldflags="-X main.Version=v1.2.0" ./cmd/main.go
var Version = "1.2.1"

func main() {
	app := bootstrap.NewApp(Version)

	if err := app.Initialize(); err != nil {
		logger.Fatal("应用初始化失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleSignals(cancel)

	go func() {
		if err := app.Start(); err != nil {
			logger.Fatal("应用启动失败: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Error("应用关闭过程中发生错误: %v", err)
	}
}

func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	<-signalChan
	logger.Info("正在安全退出程序...")
	cancel()
}
