package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Структура конфигурации агента
type agentConfig struct {
	host           string
	port           string
	reportInterval int
	pollInterval   int
}

// Конструктор конфигурации агента
func newConfig() (*agentConfig, error) {
	var err error
	config := &agentConfig{}

	// Парсинг флагов
	config.parseFlags()

	// Парсинг переменных окружения
	if err = config.parseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	return config, nil
}

// Конструктор инструкций флагов агента
func (a *agentConfig) parseFlags() {
	flag.StringVar(&a.host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.port, "port", "8080", "Port on which to listen. Example: \"8081\"")
	flag.IntVar(&a.reportInterval, "r", 10, "Metrics send interval")
	flag.IntVar(&a.pollInterval, "p", 2, "Metrics update interval")

	_ = flag.Value(a)
	flag.Var(a, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Конструктор инструкций переменных окружений агента
func (a *agentConfig) parseEnv() error {
	var err error
	if address := os.Getenv("ADDRESS"); address != "" {
		if err = a.Set(address); err != nil {
			return fmt.Errorf("error setting ADDRESS: %w", err)
		}
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		if a.reportInterval, err = strconv.Atoi(reportInterval); err != nil {
			return fmt.Errorf("error parsing REPORT_INTERVAL: %w", err)
		}
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		if a.pollInterval, err = strconv.Atoi(pollInterval); err != nil {
			return fmt.Errorf("error parsing POLL_INTERVAL: %w", err)
		}
	}

	return nil
}

// Реализация интерфейса flag.Value
func (a *agentConfig) String() string {
	return a.host + ":" + a.port
}

// Реализация интерфейса flag.Value
func (a *agentConfig) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.host = values[0]
	a.port = values[1]
	return nil
}
