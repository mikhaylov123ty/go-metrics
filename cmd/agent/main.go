package main

import (
	"log"

	"metrics/internal/client"
)

func main() {
	// Инициализация флагов агента
	config, err := NewConfig()
	if err != nil {
		log.Fatal("Build Agent Config Error:", err)
	}

	// Инициализация инстанса агента
	agentInstance := client.NewAgent(
		config.String(),
		config.PollInterval,
		config.ReportInterval,
		config.Key,
	)

	// Запуск агента
	agentInstance.Run()

}
