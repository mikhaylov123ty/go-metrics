package main

import (
	"metrics/internal/client"
)

func main() {

	// Инициализация флагов агента
	flags := buildFlags()

	// Инициализация инстанса агента
	agentInstance := client.NewAgent(flags.String(), flags.pollInterval, flags.reportInterval)

	// Запуск агента
	agentInstance.Run()

}
