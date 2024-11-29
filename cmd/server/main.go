package main

import (
	"log"

	"metrics/internal/storage/psql"

	"metrics/internal/server"
	"metrics/internal/storage"
	"metrics/pkg/logger"
)

func main() {
	// Инициализация конфигурации сервера
	config, err := NewConfig()
	if err != nil {
		log.Fatal("Build Server Config Error:", err)
	}

	// Инициализация инстанса хранения данных
	var storageInstance storage.Storage
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

		storageInstance = psqlStorage

	default:
		storageInstance = storage.NewMemoryStorage()
	}

	// Инициализация инстанса логгера
	loggerInstance, err := logger.New(config.Logger.LogLevel)
	if err != nil {
		log.Fatal("Build Logger Config Error:", err)
	}

	// Инициализация инстанса сервера
	serverInstance := server.New(
		storageInstance,
		loggerInstance,
		config.FileStorage.StoreInterval,
		config.FileStorage.FileStoragePath,
		config.FileStorage.Restore,
	)

	// Запуск сервера
	serverInstance.Start(config.String())
}
