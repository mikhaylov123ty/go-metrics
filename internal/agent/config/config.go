// Модуль config инициализирует конфигрурацию агента
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

// AgentConfig - структура конфигурации агента
type AgentConfig struct {
	Host           *Host
	ConfigFile     string
	ReportInterval float64
	PollInterval   float64
	RateLimit      int
	Key            string
	CryptoKey      string
}
type Host struct {
	Address  string
	HTTPPort string
	GRPCPort string
}

// New - конструктор конфигурации агента
func New() (*AgentConfig, error) {
	var err error
	config := &AgentConfig{Host: &Host{}}

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

	return config, nil
}

// parseFlags - Парсинг инструкций флагов агента
func (a *AgentConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&a.Host.Address, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.Host.HTTPPort, "http-port", "8080", "Port on which to listen. Example: \"8080\"")
	flag.StringVar(&a.Host.GRPCPort, "grpc-port", "", "Port on which to listen gRPC requests. Example: \"4443\"")

	// Флаги интервалов метрик
	flag.Float64Var(&a.ReportInterval, "r", 0, "Metrics send interval in seconds.")
	flag.Float64Var(&a.PollInterval, "p", 0, "Metrics update interval in seconds.")

	// Флаги подписи
	flag.StringVar(&a.Key, "k", "", "Key")

	// Флаги лимитов запросов
	flag.IntVar(&a.RateLimit, "l", 5, "Metrics simultaneously send limit. Defalut: 5")

	//Флаг публичного ключа
	flag.StringVar(&a.CryptoKey, "crypto-key", "", "Path to public cert file")

	// Флаг файла конфигурации
	flag.StringVar(&a.ConfigFile, "config", "", "Config file")

	_ = flag.Value(a.Host)
	flag.Var(a.Host, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// parseEnv - Парсинг инструкций переменных окружений агента
func (a *AgentConfig) parseEnv() error {
	if address := os.Getenv("ADDRESS"); address != "" {
		if err := a.Host.Set(address); err != nil {
			return fmt.Errorf("error setting ADDRESS: %w", err)
		}
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		interval, err := strconv.Atoi(reportInterval)
		if err != nil {
			return fmt.Errorf("invalid REPORT_INTERVAL to int conversion: %w", err)
		}

		a.ReportInterval = float64(interval)
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		interval, err := strconv.Atoi(pollInterval)
		if err != nil {
			return fmt.Errorf("invalid POLL_INTERVAL to int conversion: %w", err)
		}
		a.PollInterval = float64(interval)
	}

	if key := os.Getenv("KEY"); key != "" {
		a.Key = key
	}

	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		limit, err := strconv.Atoi(rateLimit)
		if err != nil {
			return fmt.Errorf("invalid RATE_LIMIT to int conversion: %w", err)
		}
		a.RateLimit = limit
	}

	if cert := os.Getenv("CRYPTO_KEY"); cert != "" {
		a.CryptoKey = cert
	}

	if config := os.Getenv("CONFIG"); config != "" {
		a.ConfigFile = config
	}

	if grpcPort := os.Getenv("GRPC_PORT"); grpcPort != "" {
		a.Host.GRPCPort = grpcPort
	}

	return nil
}

// initConfigFile читает и инициализирует файл конфигурации
func (a *AgentConfig) initConfigFile() error {
	fileData, err := os.ReadFile(a.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(fileData, a); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}

// UnmarshalJSON реализует интерфейс Unmarshaler
// позволяет десериализировать файл конфига с условиями
func (a *AgentConfig) UnmarshalJSON(b []byte) error {
	var err error
	var cfg struct {
		GRPCPort       string `json:"grpc_port"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if a.Host.GRPCPort == "" && cfg.GRPCPort != "" {
		a.Host.GRPCPort = cfg.GRPCPort
	}

	if a.ReportInterval == 0 && cfg.ReportInterval != "" {
		var interval time.Duration
		interval, err = time.ParseDuration(cfg.ReportInterval)
		if err != nil {
			return fmt.Errorf("error parsing report_interval: %w", err)
		}
		a.ReportInterval = interval.Seconds()
	}

	if a.PollInterval == 0 && cfg.PollInterval != "" {
		var interval time.Duration
		interval, err = time.ParseDuration(cfg.PollInterval)
		if err != nil {
			return fmt.Errorf("error parsing poll_interval: %w", err)
		}
		a.PollInterval = interval.Seconds()
	}

	if a.CryptoKey == "" && cfg.CryptoKey != "" {
		a.CryptoKey = cfg.CryptoKey
	}

	return nil
}

// String реализует интерфейс flag.Value
func (h *Host) String() string {
	return h.Address + ":" + h.HTTPPort
}

// Set реализует интерфейс flag.Value
func (h *Host) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	h.Address = values[0]
	h.HTTPPort = values[1]

	return nil
}
