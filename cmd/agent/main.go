package main

import (
	"metrics/internal/client"
)

func main() {

	// Инициализация инстанса агента
	agentInstance := client.NewAgent("http://localhost:8080/")

	// Запуск агента
	agentInstance.Run()

}
