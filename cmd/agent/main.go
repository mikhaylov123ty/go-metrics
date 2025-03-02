package main

import (
	"log"
	"net/http"
	
	_ "net/http/pprof"

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
	go agentInstance.Run()

	http.ListenAndServe(":30012", nil)
}
