package config

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentConfigParseFlagSetRateLimit(t *testing.T) {
	cfg := &AgentConfig{
		ServerAddress:  "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      1,
	}

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	err := cfg.parseFlagSet(fs, []string{"-a=localhost:9090", "-p=1", "-r=5", "-l=7"})
	require.NoError(t, err)

	assert.Equal(t, "localhost:9090", cfg.ServerAddress)
	assert.Equal(t, 1*time.Second, cfg.PollInterval)
	assert.Equal(t, 5*time.Second, cfg.ReportInterval)
	assert.Equal(t, 7, cfg.RateLimit)
}

func TestAgentConfigValidateRateLimit(t *testing.T) {
	cfg := &AgentConfig{
		ServerAddress:  "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      0,
	}

	err := cfg.validate()
	require.Error(t, err)
	assert.ErrorContains(t, err, "rate limit must be positive")
}
