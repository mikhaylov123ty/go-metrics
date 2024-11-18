package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Структура конфигурации сервера
type serverConfig struct {
	host            string
	port            string
	logLevel        string
	storeInterval   int
	fileStoragePath string
	restore         bool
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
	flag.IntVar(&s.storeInterval, "i", 300, "Interval in seconds, to store metrics in file.")
	flag.StringVar(&s.fileStoragePath, "f", "tempFile.txt", "Path to file to store metrics. Example: ./tempFile.txt")
	flag.BoolVar(&s.restore, "r", true, "Restore previous metrics from file.")

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

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		interval, err := strconv.Atoi(storeInterval)
		if err != nil {
			return fmt.Errorf("error parsing STORE_INTERVAL: %w", err)
		}
		s.storeInterval = interval
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		s.fileStoragePath = fileStoragePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		if restore == "true" {
			s.restore = true
		}
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
