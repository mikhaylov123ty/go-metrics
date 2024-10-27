package main

import (
	"log"

	"metrics/internal/server"
	"metrics/internal/storage"
)

func main() {
	// Инициализация конфигурации сервера
	config, err := newConfig()
	if err != nil {
		log.Fatal("Build Server Config Error:", err)
	}

	//Инициализация инстанса хранения данных
	storageInstance := storage.NewMemoryStorage()

	// Инициализация инстанса сервера
	serverInstance := server.New(&storageInstance)

	// Запуск сервера
	serverInstance.Start(config.String())
}
