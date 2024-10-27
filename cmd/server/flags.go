package main

import (
	"flag"
	"fmt"
	"strings"
)

// Структура флагов сервера
type serverFlags struct {
	host string
	port string
}

// Конструктор структуры флагов сервера
func buildFlags() *serverFlags {
	flags := &serverFlags{}

	flags.parseFlags()

	return flags
}

// Конструктор инструкций флагов сервера
func (sf *serverFlags) parseFlags() {
	flag.StringVar(&sf.host, "h", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&sf.port, "p", "8080", "Port on which to listen. Example: \"8081\"")

	_ = flag.Value(sf)
	flag.Var(sf, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Реализация интерфейса flag.Value
func (sf *serverFlags) String() string {
	return sf.host + ":" + sf.port
}

// Реализация интерфейса flag.Value
func (sf *serverFlags) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	sf.host = values[0]
	sf.port = values[1]
	return nil
}
