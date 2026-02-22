package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

// Создаем новый экземпляр конфигурации сервера, загружая значения конфигурации
// Приоритет загрузки: переменные окружения > флаги > значения по умолчанию
func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Address:  ":8080",
		LogLevel: "info",
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
	return cfg
}

func (sc *ServerConfig) parseFlags() error {
	flag.StringVar(&sc.Address, "a", sc.Address, "Server address")
	flag.StringVar(&sc.LogLevel, "loglvl", sc.LogLevel, "Log level")
	flag.Parse()

	// Проверяем, что не переданы неизвестные флаги
	if flag.NArg() > 0 {
		flag.Usage()
		return fmt.Errorf("unknown flags or arguments: %v", flag.Args())
	}
	return nil

}

func (sc *ServerConfig) envParse() error {
	err := env.Parse(sc)
	if err != nil {
		return err
	}
	return nil
}

func (sc *ServerConfig) validate() error {
	var errs []error
	if sc.Address == "" {
		errs = append(errs, fmt.Errorf("server address must be set, got empty value"))
	}
	if sc.LogLevel == "" {
		errs = append(errs, fmt.Errorf("server log level must be set, got empty value"))
	}
	return errors.Join(errs...)
}
