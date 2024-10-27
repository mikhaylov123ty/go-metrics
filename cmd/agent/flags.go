package main

import (
	"flag"
	"fmt"
	"strings"
)

// Структура флагов агента
type agentFlags struct {
	host           string
	port           string
	reportInterval int
	pollInterval   int
}

// Конструктор структуры флагов агента
func buildFlags() *agentFlags {
	flags := &agentFlags{}

	flags.parseFlags()

	return flags
}

// Конструктор инструкций флагов агента
func (af *agentFlags) parseFlags() {
	flag.StringVar(&af.host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&af.port, "port", "8080", "Port on which to listen. Example: \"8081\"")
	flag.IntVar(&af.reportInterval, "r", 10, "Metrics send interval")
	flag.IntVar(&af.pollInterval, "p", 2, "Metrics update interval")

	_ = flag.Value(af)
	flag.Var(af, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Реализация интерфейса flag.Value
func (af *agentFlags) String() string {
	return af.host + ":" + af.port
}

// Реализация интерфейса flag.Value
func (af *agentFlags) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}
	af.host = values[0]
	af.port = values[1]
	return nil
}
