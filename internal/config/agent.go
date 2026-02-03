package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
}

func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{
		ServerAddress:  "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
	cfg.parseFlags()
	return cfg
}

func NewTestAgentConfig(serverAddress string) *AgentConfig {
	return &AgentConfig{
		ServerAddress:  serverAddress,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}

func (ac *AgentConfig) parseFlags() {
	// Флаг для адреса сервера
	flag.StringVar(&ac.ServerAddress, "a", ac.ServerAddress, "Server address")

	// Флаги для интервалов времени
	var pollIntervalSec, reportIntervalSec float64
	flag.Float64Var(&pollIntervalSec, "p", 2.0, "Poll interval in seconds")
	flag.Float64Var(&reportIntervalSec, "r", 10.0, "Report interval in seconds")

	flag.Parse()

	// Конвертируем в time.Duration
	ac.PollInterval = time.Duration(pollIntervalSec * float64(time.Second))
	ac.ReportInterval = time.Duration(reportIntervalSec * float64(time.Second))

	// Проверяем, что не переданы неизвестные флаги
	if flag.NArg() > 0 {
		fmt.Printf("Error: unknown flags or arguments: %v\n", flag.Args())
		flag.Usage()
		os.Exit(1)
	}

	// Валидация значений
	if ac.PollInterval <= 0 {
		fmt.Printf("Error: poll interval must be positive, got: %.1f seconds\n", pollIntervalSec)
		os.Exit(1)
	}

	if ac.ReportInterval <= 0 {
		fmt.Printf("Error: report interval must be positive, got: %.1f seconds\n", reportIntervalSec)
		os.Exit(1)
	}

}
