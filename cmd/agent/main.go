package main

import (
	"log"

	"metrics/internal/client"
	"metrics/internal/client/config"
)

func main() {
	// Инициализация флагов агента
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Build Agent Config Error:", err)
	}

	// Инициализация инстанса агента
	agentInstance := client.NewAgent(cfg)

	// Запуск агента
	agentInstance.Run()
}
