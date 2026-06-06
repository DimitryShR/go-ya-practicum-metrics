package config

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfigParseFlagSet(t *testing.T) {
	cfg := &ServerConfig{
		Address:         ":8080",
		LogLevel:        "info",
		StoreInterval:   300 * time.Second,
		FileStoragePath: "",
		Restore:         true,
		SignKey:         "",
		AuditFile:       "",
		AuditURL:        "",
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	err := cfg.parseFlagSet(fs, []string{
		"-a=localhost:9090",
		"-loglvl=debug",
		"-i=12.5",
		"-f=/tmp/metrics-db.json",
		"-r=false",
		"-d=postgres://app:secret@localhost:5432/metrics?sslmode=disable",
		"-m=postgres://migrator:secret@localhost:5433/metrics_migrate?sslmode=require",
		"-k=test-sign-key",
		"-audit-file=/var/log/audit.json",
		"-audit-url=http://audit.example.com/events",
	})
	require.NoError(t, err)

	assert.Equal(t, "localhost:9090", cfg.Address)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 12500*time.Millisecond, cfg.StoreInterval)
	assert.Equal(t, "/tmp/metrics-db.json", cfg.FileStoragePath)
	assert.False(t, cfg.Restore)
	assert.Equal(t, "test-sign-key", cfg.SignKey)
	assert.Equal(t, "/var/log/audit.json", cfg.AuditFile)
	assert.Equal(t, "http://audit.example.com/events", cfg.AuditURL)

	assert.Equal(
		t,
		"host=localhost port=5432 user=app password=secret dbname=metrics sslmode=disable",
		cfg.DBDsn.GetKeywordDSN(),
	)
	assert.Equal(
		t,
		"postgres://migrator:secret@localhost:5433/metrics_migrate?sslmode=require",
		cfg.DBMigrateDsn.GetURL(),
	)
}

func TestServerConfigAuditDefaultsEmpty(t *testing.T) {
	cfg := &ServerConfig{
		Address:       ":8080",
		LogLevel:      "info",
		StoreInterval: 300 * time.Second,
		Restore:       true,
		AuditFile:     "",
		AuditURL:      "",
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	err := cfg.parseFlagSet(fs, []string{})
	require.NoError(t, err)

	assert.Empty(t, cfg.AuditFile)
	assert.Empty(t, cfg.AuditURL)
}

func TestServerConfigParseFlagSetInvalidDSN(t *testing.T) {
	cfg := &ServerConfig{
		Address:       ":8080",
		LogLevel:      "info",
		StoreInterval: 300 * time.Second,
		Restore:       true,
		AuditFile:     "",
		AuditURL:      "",
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	err := cfg.parseFlagSet(fs, []string{"-d=not-a-valid-dsn"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "dsn must be URL format")
}
