// Модуль config инициализирует конфигрурацию сервера
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ServerConfig - структура конфигурации сервера
type ServerConfig struct {
	Host           string
	Port           string
	Logger         *Logger
	FileStorage    *FileStorage
	DB             *DB
	Key            string
	PrivateKeyFile string
}

// Logger - структура конфигруации логгера
type Logger struct {
	LogLevel string
}

// FileStorage - структура конфигурации хранилища
type FileStorage struct {
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

// DB - структура конфигруации БД
type DB struct {
	Address string
}

// New - конструктор конфигурации сервера
func New() (*ServerConfig, error) {
	var err error
	config := &ServerConfig{Logger: &Logger{}, FileStorage: &FileStorage{}, DB: &DB{}}

	// Парсинг флагов
	config.parseFlags()

	// Парсинг переменных окружения
	if err = config.parseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	if config.DB.Address != "" {
		config.FileStorage.Restore = false
	}

	return config, nil
}

// Парсинг инструкций флагов сервера
func (s *ServerConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&s.Host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&s.Port, "port", "8080", "Port on which to listen. Example: \"8080\"")

	// Флаги логирования
	flag.StringVar(&s.Logger.LogLevel, "l", "info", "Log level. Example: \"info\"")

	// Флаги файлового хранилища
	flag.IntVar(&s.FileStorage.StoreInterval, "i", 10, "Interval in seconds, to store metrics in file.")
	flag.StringVar(&s.FileStorage.FileStoragePath, "f", "tempFile.txt", "Path to file to store metrics. Example: ./tempFile.txt")
	flag.BoolVar(&s.FileStorage.Restore, "r", true, "Restore previous metrics from file.")

	// Флаги БД
	flag.StringVar(&s.DB.Address, "d", "", "Host which to connect to DB. Example: \"postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable\"")

	// Флаги подписи и шифрования
	flag.StringVar(&s.Key, "k", "", "Key")

	// Флаги приватного и публичного ключей
	flag.StringVar(&s.PrivateKeyFile, "crypto-key", "", "Path to private crypto key file")

	_ = flag.Value(s)
	flag.Var(s, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Парсинг инструкций переменных окружений сервера
func (s *ServerConfig) parseEnv() error {
	var err error
	if address := os.Getenv("ADDRESS"); address != "" {
		if err = s.Set(address); err != nil {
			return err
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		s.Logger.LogLevel = logLevel
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		interval, err := strconv.Atoi(storeInterval)
		if err != nil {
			return fmt.Errorf("error parsing STORE_INTERVAL: %w", err)
		}
		s.FileStorage.StoreInterval = interval
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		s.FileStorage.FileStoragePath = fileStoragePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		if restore == "true" {
			s.FileStorage.Restore = true
		}
	}

	if address := os.Getenv("DATABASE_DSN"); address != "" {
		s.DB.Address = address
	}

	if key := os.Getenv("KEY"); key != "" {
		s.Key = key
	}

	if privateKey := os.Getenv("CRYPTO_KEY"); privateKey != "" {
		s.PrivateKeyFile = privateKey
	}

	return nil
}

// String реализаует интерфейс flag.Value
func (s *ServerConfig) String() string {
	return s.Host + ":" + s.Port
}

// Set реализует интерфейса flag.Value
func (s *ServerConfig) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	s.Host = values[0]
	s.Port = values[1]

	return nil
}
