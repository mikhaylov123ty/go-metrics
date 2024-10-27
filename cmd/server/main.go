package main

import (
	"metrics/internal/server"
	"metrics/internal/storage"
)

func main() {
	//Инициализация инстанса хранения данных
	storageInstance := storage.NewMemoryStorage()

	// Инициализация инстанса сервера
	serverInstance := server.New(&storageInstance)

	// Инициализация флагов сервера
	flags := buildFlags()

	// Запуск сервера
	serverInstance.Start(flags.String())
}
