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
	Host           string
	Port           string
	ReportInterval float64
	PollInterval   float64
	Key            string
	RateLimit      int
	CryptoKey      string
	ConfigFile     string
}

// New - конструктор конфигурации агента
func New() (*AgentConfig, error) {
	var err error
	config := &AgentConfig{}

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
	flag.StringVar(&a.Host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.Port, "port", "8080", "Port on which to listen. Example: \"8081\"")

	// Флаги интервалов метрик
	flag.Float64Var(&a.ReportInterval, "r", 0, "Metrics send interval in seconds.")
	flag.Float64Var(&a.PollInterval, "p", 0, "Metrics update interval in seconds.")

	// Флаги подписи и шифрования
	flag.StringVar(&a.Key, "k", "", "Key")

	// Флаги лимитов запросов
	flag.IntVar(&a.RateLimit, "l", 5, "Metrics simultaneously send limit. Defalut: 5")

	//Флаг публичного ключа
	flag.StringVar(&a.CryptoKey, "crypto-key", "", "Path to public cert file")

	// Флаг файла конфигурации
	flag.StringVar(&a.ConfigFile, "config", "", "Config file")

	_ = flag.Value(a)
	flag.Var(a, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// parseEnv - Парсинг инструкций переменных окружений агента
func (a *AgentConfig) parseEnv() error {
	var err error
	if address := os.Getenv("ADDRESS"); address != "" {
		if err = a.Set(address); err != nil {
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
		if a.RateLimit, err = strconv.Atoi(rateLimit); err != nil {
			return fmt.Errorf("error parsing RATE_LIMIT: %w", err)
		}
	}

	if cert := os.Getenv("CRYPTO_KEY"); cert != "" {
		a.CryptoKey = cert
	}

	if config := os.Getenv("CONFIG"); config != "" {
		a.ConfigFile = config
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
		Address        string `json:"address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if (a.Host == "" && a.Port == "") && cfg.Address != "" {
		if err = a.Set(cfg.Address); err != nil {
			return fmt.Errorf("error parsing address: %w", err)
		}
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
func (a *AgentConfig) String() string {
	return a.Host + ":" + a.Port
}

// Set реализует интерфейс flag.Value
func (a *AgentConfig) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.Host = values[0]
	a.Port = values[1]

	return nil
}
