package main

import (
	"fmt"
	"log"
	"net/http"

	_ "net/http/pprof"

	"metrics/internal/client"
	"metrics/internal/client/config"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Agent Build Version: %s\n", buildVersion)
	fmt.Printf("Agent Build Date: %s\n", buildDate)
	fmt.Printf("Agent Build Commit: %s\n", buildCommit)

	// Инициализация флагов агента
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Build Agent Config Error:", err)
	}

	// Инициализация инстанса агента
	agentInstance := client.NewAgent(cfg)

	// Запуск агента
	go agentInstance.Run()

	if err = http.ListenAndServe(":30012", nil); err != nil {
		log.Fatal("HTTP Server Error:", err)
	}
}
