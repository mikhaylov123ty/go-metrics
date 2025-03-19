package main

import (
	"context"
	"fmt"
	"log"
	"metrics/internal/client"
	"metrics/internal/client/config"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"sync"
	"syscall"
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

	// Создание контекста с сигналами
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	defer stop()

	// Создание группы ожидания
	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Запуск агента
	go agentInstance.Run(ctx, wg)

	// Запуск сервера профилирования
	srv := http.Server{Addr: ":30012"}
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP Server Error:", err)
		}
	}()

	// Ожидание сигнала
	<-ctx.Done()

	// Остановка сервера
	if err = srv.Shutdown(ctx); err != nil && err != context.Canceled {
		log.Fatal("HTTP Server Shutdown Failed:", err)
	}

	// Ожидание завершения горутин
	wg.Wait()

	log.Println("Agent Shutdown gracefully")
}
