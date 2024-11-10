package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Структура конфигурации сервера
type serverConfig struct {
	host     string
	port     string
	logLevel string
}

// Конструктор конфигурации сервера
func newConfig() (*serverConfig, error) {
	var err error
	config := &serverConfig{}

	// Парсинг флагов
	config.parseFlags()

	// Парсинг переменных окружения
	if err = config.parseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	return config, nil
}

// Конструктор инструкций флагов сервера
func (s *serverConfig) parseFlags() {
	flag.StringVar(&s.host, "h", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&s.port, "p", "8080", "Port on which to listen. Example: \"8081\"")
	flag.StringVar(&s.logLevel, "l", "info", "Log level. Example: \"info\"")

	_ = flag.Value(s)
	flag.Var(s, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Конструктор инструкций переменных окружений сервера
func (s *serverConfig) parseEnv() error {
	var err error
	if address := os.Getenv("ADDRESS"); address != "" {
		if err = s.Set(address); err != nil {
			return err
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		s.logLevel = logLevel
	}

	return nil
}

// Реализация интерфейса flag.Value
func (s *serverConfig) String() string {
	return s.host + ":" + s.port
}

// Реализация интерфейса flag.Value
func (s *serverConfig) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	s.host = values[0]
	s.port = values[1]
	return nil
}
