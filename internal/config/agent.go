package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
	SignKey        string
	RateLimit      int
}

// Создаем новый экземпляр конфигурации агента, загружая значения конфигурации
// Приоритет загрузки: переменные окружения > флаги > значения по умолчанию
func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{
		ServerAddress:  "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		SignKey:        "",
		RateLimit:      1,
	}
	if err := cfg.parseFlags(); err != nil {
		fmt.Println("config flags parse error:", err)
		os.Exit(1)
	}
	if err := cfg.envParse(); err != nil {
		fmt.Println("config env parse error:", err)
		os.Exit(1)
	}
	if err := cfg.validate(); err != nil {
		fmt.Println("config validation error:", err)
		os.Exit(1)
	}
	cfg.normalizeAddress()
	return cfg
}

func NewTestAgentConfig(serverAddress string) *AgentConfig {
	return &AgentConfig{
		ServerAddress:  serverAddress,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      1,
	}
}

// Добавляем схему по умолчанию, если она не указана в адресе сервера
func (ac *AgentConfig) normalizeAddress() {
	if !strings.Contains(ac.ServerAddress, "://") {
		ac.ServerAddress = "http://" + ac.ServerAddress
	}
}

// Извлекаем конфигурацию из переменных окружения и вносим изменения в структуру конфигурации
func (ac *AgentConfig) envParse() error {
	tmpCfg := struct {
		PollInterval   *float64 `env:"POLL_INTERVAL"`
		ReportInterval *float64 `env:"REPORT_INTERVAL"`
		ServerAddress  *string  `env:"ADDRESS"`
		SignKey        *string  `env:"KEY"`
		RateLimit      *int     `env:"RATE_LIMIT"`
	}{}
	err := env.Parse(&tmpCfg)
	if err != nil {
		return err
	}
	// Проверяем, что переменные окружения не пустые и конвертируем в нужные типы
	if tmpCfg.PollInterval != nil {
		ac.PollInterval = time.Duration(*tmpCfg.PollInterval * float64(time.Second))
	}
	if tmpCfg.ReportInterval != nil {
		ac.ReportInterval = time.Duration(*tmpCfg.ReportInterval * float64(time.Second))
	}
	if tmpCfg.ServerAddress != nil {
		ac.ServerAddress = *tmpCfg.ServerAddress
	}
	if tmpCfg.SignKey != nil {
		ac.SignKey = *tmpCfg.SignKey
	}
	if tmpCfg.RateLimit != nil {
		ac.RateLimit = *tmpCfg.RateLimit
	}
	return nil
}

// Парсим флаги командной строки и вносим изменения в структуру конфигурации
func (ac *AgentConfig) parseFlags() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	// Подавляем вывод
	fs.SetOutput(io.Discard)

	return ac.parseFlagSet(fs, os.Args[1:])
}

func (ac *AgentConfig) parseFlagSet(fs *flag.FlagSet, args []string) error {
	// Флаг для адреса сервера
	fs.StringVar(&ac.ServerAddress, "a", ac.ServerAddress, "Server address")
	fs.StringVar(&ac.SignKey, "k", ac.SignKey, "Key for sign data")
	// Флаги для интервалов времени
	var pollIntervalSec, reportIntervalSec float64
	fs.Float64Var(&pollIntervalSec, "p", ac.PollInterval.Seconds(), "Poll interval in seconds")
	fs.Float64Var(&reportIntervalSec, "r", ac.ReportInterval.Seconds(), "Report interval in seconds")
	fs.IntVar(&ac.RateLimit, "l", ac.RateLimit, "Maximum number of concurrent outbound requests")

	// Парсим флаги
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Конвертируем в time.Duration
	ac.PollInterval = time.Duration(pollIntervalSec * float64(time.Second))
	ac.ReportInterval = time.Duration(reportIntervalSec * float64(time.Second))

	// Проверяем, что не переданы неизвестные флаги
	if fs.NArg() > 0 {
		return fmt.Errorf("unknown flags or arguments: %v", fs.Args())
	}
	return nil
}

// Валидация конфигурации с накоплением ошибок
func (ac *AgentConfig) validate() error {
	var errs []error
	if ac.PollInterval <= 0 {
		errs = append(errs, fmt.Errorf("poll interval must be positive, got: %s", ac.PollInterval))
	}
	if ac.ReportInterval <= 0 {
		errs = append(errs, fmt.Errorf("report interval must be positive, got: %s", ac.ReportInterval))
	}
	if ac.ServerAddress == "" {
		errs = append(errs, fmt.Errorf("server address must be set, got empty value"))
	}
	if ac.RateLimit <= 0 {
		errs = append(errs, fmt.Errorf("rate limit must be positive, got: %d", ac.RateLimit))
	}
	return errors.Join(errs...)
}

func (ac *AgentConfig) String() string {
	return fmt.Sprintf(
		"Server address: %s; Poll interval: %s; Report interval: %s; Sign key: %s; Rate limit: %d",
		ac.ServerAddress, ac.PollInterval, ac.ReportInterval, ac.SignKey, ac.RateLimit,
	)
}
