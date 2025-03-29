package main

import (
	"context"
	"fmt"
	"log"
	"metrics/internal/server/gRPC"
	"os/signal"
	"syscall"

	"metrics/internal/server"
	"metrics/internal/server/api"
	"metrics/internal/server/config"
	"metrics/internal/server/metrics"
	"metrics/internal/storage/memory"
	"metrics/internal/storage/psql"
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

	// Инициализация инстанса хранения данных
	var apiStorageCommands *api.StorageCommands
	var gRPCStorageCommands *gRPC.StorageCommands
	var metricsFileStorage *metrics.MetricsFileStorage
	switch {
	case cfg.DB.Address != "":
		var psqlStorage *psql.DataBase
		psqlStorage, err = psql.NewPSQLDataBase(
			cfg.DB.Address,
		)
		if err != nil {
			log.Fatal("Build Server Storage Connection Error:", err)
		}

		defer func() {
			if err = psqlStorage.Instance.Close(); err != nil {
				loggerInstance.Errorf("Build Server Storage Close Instance Error: %s", err.Error())
			}
		}()

		if err = psqlStorage.BootStrap(cfg.DB.Address); err != nil {
			log.Fatal("Build Server Storage Bootstrap Error:", err)
		}

		apiStorageCommands = api.NewStorageService(
			psqlStorage,
			psqlStorage,
			psqlStorage,
		)

		gRPCStorageCommands = gRPC.NewStorageService(
			psqlStorage,
			psqlStorage,
		)

		metricsFileStorage = metrics.NewMetricsFileStorage(psqlStorage, cfg.FileStorage.FileStoragePath)
		loggerInstance.Info("Storage: postgres")

	default:
		memStorage := memory.NewMemoryStorage()

		apiStorageCommands = api.NewStorageService(
			memStorage,
			memStorage,
			nil,
		)
		gRPCStorageCommands = gRPC.NewStorageService(
			memStorage,
			memStorage,
		)

		metricsFileStorage = metrics.NewMetricsFileStorage(memStorage, cfg.FileStorage.FileStoragePath)

		loggerInstance.Info("Storage: memory")
	}

	// Инициализация инстанса сервера
	serverInstance := server.New(
		apiStorageCommands,
		gRPCStorageCommands,
		metricsFileStorage,
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
