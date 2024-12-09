package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Структура конфигурации агента
type AgentConfig struct {
	Host           string
	Port           string
	ReportInterval int
	PollInterval   int
	Key            string
}

// Конструктор конфигурации агента
func NewConfig() (*AgentConfig, error) {
	var err error
	config := &AgentConfig{}

	// Парсинг флагов
	config.parseFlags()

	// Парсинг переменных окружения
	if err = config.parseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	return config, nil
}

// Конструктор инструкций флагов агента
func (a *AgentConfig) parseFlags() {
	flag.StringVar(&a.Host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.Port, "port", "8080", "Port on which to listen. Example: \"8081\"")

	flag.IntVar(&a.ReportInterval, "r", 10, "Metrics send interval")
	flag.IntVar(&a.PollInterval, "p", 2, "Metrics update interval")

	flag.StringVar(&a.Key, "k", "", "Key")

	_ = flag.Value(a)
	flag.Var(a, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Конструктор инструкций переменных окружений агента
func (a *AgentConfig) parseEnv() error {
	var err error
	if address := os.Getenv("ADDRESS"); address != "" {
		if err = a.Set(address); err != nil {
			return fmt.Errorf("error setting ADDRESS: %w", err)
		}
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		if a.ReportInterval, err = strconv.Atoi(reportInterval); err != nil {
			return fmt.Errorf("error parsing REPORT_INTERVAL: %w", err)
		}
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		if a.PollInterval, err = strconv.Atoi(pollInterval); err != nil {
			return fmt.Errorf("error parsing POLL_INTERVAL: %w", err)
		}
	}

	if key := os.Getenv("KEY"); key != "" {
		a.Key = key
	}

	return nil
}

// Реализация интерфейса flag.Value
func (a *AgentConfig) String() string {
	return a.Host + ":" + a.Port
}

// Реализация интерфейса flag.Value
func (a *AgentConfig) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.Host = values[0]
	a.Port = values[1]
	return nil
}
