// Модуль config инициализирует конфигрурацию сервера
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ServerConfig - структура конфигурации сервера
type ServerConfig struct {
	Host        string
	Port        string
	Logger      *Logger
	FileStorage *FileStorage
	DB          *DB
	Key         string
	CryptoKey   string
	ConfigFile  string
}

// Logger - структура конфигруации логгера
type Logger struct {
	LogLevel string
}

// FileStorage - структура конфигурации хранилища
type FileStorage struct {
	StoreInterval   float64
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

	// Инициализация конфига из файла
	if config.ConfigFile != "" {
		if err = config.initConfigFile(); err != nil {
			return nil, fmt.Errorf("failed init config file: %w", err)
		}
	}

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
	flag.Float64Var(&s.FileStorage.StoreInterval, "i", 0, "Interval in seconds, to store metrics in file.")
	flag.StringVar(&s.FileStorage.FileStoragePath, "f", "", "Path to file to store metrics. Example: ./tempFile.txt")
	flag.BoolVar(&s.FileStorage.Restore, "r", true, "Restore previous metrics from file.")

	// Флаги БД
	flag.StringVar(&s.DB.Address, "d", "", "Host which to connect to DB. Example: \"postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable\"")

	// Флаги подписи и шифрования
	flag.StringVar(&s.Key, "k", "", "Key")

	// Флаги приватного и публичного ключей
	flag.StringVar(&s.CryptoKey, "crypto-key", "", "Path to private crypto key file")

	// Флаг файла конфигурации
	flag.StringVar(&s.ConfigFile, "config", "", "Config file")

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
			return fmt.Errorf("invalid STORE_INTERVAL to int conversion: %w", err)
		}
		s.FileStorage.StoreInterval = float64(interval)
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
		s.CryptoKey = privateKey
	}

	if config := os.Getenv("CONFIG"); config != "" {
		s.ConfigFile = config
	}

	return nil
}

// initConfigFile читает и инициализирует файл конфигурации
func (s *ServerConfig) initConfigFile() error {
	fileData, err := os.ReadFile(s.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(fileData, s); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}

// UnmarshalJSON реализует интерфейс Unmarshaler
// позволяет десериализировать файл конфига с условиями
func (s *ServerConfig) UnmarshalJSON(b []byte) error {
	var err error
	var cfg struct {
		Address       string `json:"address"`
		Restore       bool   `json:"restore"`
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		DatabaseDSN   string `json:"database_dsn"`
		CryptoKey     string `json:"crypto_key"`
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if (s.Host == "" && s.Port == "") && cfg.Address != "" {
		if err = s.Set(cfg.Address); err != nil {
			return fmt.Errorf("error parsing address: %w", err)
		}
	}

	if !s.FileStorage.Restore && cfg.Restore {
		s.FileStorage.Restore = true
	}

	if s.FileStorage.StoreInterval == 0 && cfg.StoreInterval != "" {
		var interval time.Duration
		interval, err = time.ParseDuration(cfg.StoreInterval)
		if err != nil {
			return fmt.Errorf("error parsing store_interval: %w", err)
		}
		s.FileStorage.StoreInterval = interval.Seconds()
	}

	if s.FileStorage.FileStoragePath == "" && cfg.StoreFile != "" {
		s.FileStorage.FileStoragePath = cfg.StoreFile
	}

	if s.DB.Address == "" && cfg.DatabaseDSN != "" {
		s.DB.Address = cfg.DatabaseDSN
	}

	if s.CryptoKey == "" && cfg.CryptoKey != "" {
		s.CryptoKey = cfg.CryptoKey
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
