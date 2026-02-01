package config

import (
	"time"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
}

func NewDefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://localhost:8080",
	}
}

func NewCustomServerAddressAgentConfig(serverAddress string) *AgentConfig {
	return &AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  serverAddress,
	}
}
