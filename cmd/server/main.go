package main

import (
	"log"
	"metrics/internal/storage"

	"metrics/internal/server"
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
		psqlStorage, err := storage.NewPSQLDataBase(
			config.DB.Address,
		)
		if err != nil {
			log.Fatal("Build Server Storage Error:", err)
		}
		defer psqlStorage.DB.Close()

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
