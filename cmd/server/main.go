package main

import (
	"log"

	"metrics/internal/server"
	"metrics/internal/server/api"
	"metrics/internal/server/config"
	"metrics/internal/server/metrics"
	"metrics/internal/storage/memory"
	"metrics/internal/storage/psql"
	"metrics/pkg/logger"
)

func main() {
	// Инициализация конфигурации сервера
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Build Server Config Error:", err)
	}

	// Инициализация инстанса хранения данных
	var storageCommands *api.StorageCommands
	switch {
	case cfg.DB.Address != "":
		psqlStorage, err := psql.NewPSQLDataBase(
			cfg.DB.Address,
		)
		if err != nil {
			log.Fatal("Build Server Storage Connection Error:", err)
		}
		defer psqlStorage.Instance.Close()

		if err = psqlStorage.BootStrap(cfg.DB.Address); err != nil {
			log.Fatal("Build Server Storage Bootstrap Error:", err)
		}

		storageCommands = api.NewStorageService(
			psqlStorage,
			psqlStorage,
			psqlStorage,
			psqlStorage,
			psqlStorage,
		)
		log.Println("Storage: postgres")

	default:
		memStorage := memory.NewMemoryStorage()

		storageCommands = api.NewStorageService(
			memStorage,
			memStorage,
			memStorage,
			memStorage,
			nil,
		)
		log.Println("Storage: memory")
	}

	// Инициализация инстанса логгера
	loggerInstance, err := logger.New(cfg.Logger.LogLevel)
	if err != nil {
		log.Fatal("Build Logger Config Error:", err)
	}

	metricsFileStorage := metrics.NewMetricsFileStorage(storageCommands, cfg.FileStorage.FileStoragePath)

	// Инициализация инстанса сервера
	serverInstance := server.New(
		storageCommands,
		metricsFileStorage,
		loggerInstance,
		cfg,
	)

	// Запуск сервера
	serverInstance.Start(cfg.String())
}
