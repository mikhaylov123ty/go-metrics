package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"metrics/internal/storage"

	"metrics/internal/server"
	"metrics/internal/server/config"
	"metrics/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Server Build Version: %s\n", buildVersion)
	fmt.Printf("Server Build Date: %s\n", buildDate)
	fmt.Printf("Server Build Commit: %s\n", buildCommit)

	// Инициализация конфигурации сервера
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Build Server Config Error:", err)
	}

	// Инициализация инстанса логгера
	loggerInstance, err := logger.New(cfg.Logger.LogLevel)
	if err != nil {
		log.Fatal("Build Logger Config Error:", err)
	}

	loggerInstance.Debugf("Config: %+v\n", *cfg)

	// Инициализация инстанса сервера
	storageInstance := storage.NewStorage(cfg.DB.Address, cfg.FileStorage.FileStoragePath)
	defer storageInstance.Closer()

	serverInstance := server.New(
		storageInstance.ApiStorageCommands,
		storageInstance.GRPCStorageCommands,
		storageInstance.MetricsFileStorage,
		loggerInstance,
		cfg,
	)

	// Создание контекса с сигналами
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	defer stop()

	// Запуск сервера
	if err = serverInstance.Start(ctx, cfg.Host); err != nil {
		log.Fatal("Build Server Start Error: ", err)
	}

	loggerInstance.Warn("Server Shutdown gracefully")
}
