package main

import (
	"log"

	"metrics/internal/server"
	"metrics/internal/server/api"
	"metrics/internal/storage/memory"
	"metrics/internal/storage/psql"
	"metrics/pkg/logger"
)

func main() {
	// Инициализация конфигурации сервера
	config, err := NewConfig()
	if err != nil {
		log.Fatal("Build Server Config Error:", err)
	}

	// Инициализация инстанса хранения данных
	var storageCommands *api.StorageCommands
	switch {
	case config.DB.Address != "":
		psqlStorage, err := psql.NewPSQLDataBase(
			config.DB.Address,
		)
		if err != nil {
			log.Fatal("Build Server Storage Connection Error:", err)
		}
		defer psqlStorage.Instance.Close()

		if err = psqlStorage.BootStrap(config.DB.Address); err != nil {
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
	loggerInstance, err := logger.New(config.Logger.LogLevel)
	if err != nil {
		log.Fatal("Build Logger Config Error:", err)
	}

	// Инициализация инстанса сервера
	serverInstance := server.New(
		storageCommands,
		loggerInstance,
		config.FileStorage.StoreInterval,
		config.FileStorage.FileStoragePath,
		config.FileStorage.Restore,
	)

	// Запуск сервера
	serverInstance.Start(config.String())
}
