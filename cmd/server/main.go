package main

import (
	"log"

	"metrics/internal/server"
	"metrics/internal/storage"
	"metrics/pkg/logger"
)

func main() {
	// Инициализация конфигурации сервера
	config, err := newConfig()
	if err != nil {
		log.Fatal("Build Server Config Error:", err)
	}

	//Инициализация инстанса хранения данных
	storageInstance := storage.NewMemoryStorage()

	loggerInstance, err := logger.New(config.logLevel)
	if err != nil {
		log.Fatal("Build Logger Config Error:", err)
	}

	// Инициализация инстанса сервера
	serverInstance := server.New(&storageInstance, loggerInstance)

	// Запуск сервера
	serverInstance.Start(config.String())
}
