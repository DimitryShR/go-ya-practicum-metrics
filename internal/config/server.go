package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config/db"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address         string        // `env:"ADDRESS"`
	LogLevel        string        // `env:"LOG_LEVEL"`
	StoreInterval   time.Duration // `env:"STORE_INTERVAL"`
	FileStoragePath string        // `env:"FILE_STORAGE_PATH"`
	Restore         bool          // `env:"RESTORE"`
	DbDsn           db.PgConn     // `env:"DATABASE_DSN"`
}

// Создаем новый экземпляр конфигурации сервера, загружая значения конфигурации
// Приоритет загрузки: переменные окружения > флаги > значения по умолчанию
func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Address:         ":8080",
		LogLevel:        "info",
		StoreInterval:   300 * time.Second,
		FileStoragePath: "/tmp/metrics-db.json",
		Restore:         true,
		DbDsn:           db.PgConn{},
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
	flag.StringVar(&sc.FileStoragePath, "f", sc.FileStoragePath, "File storage path")
	flag.BoolVar(&sc.Restore, "r", sc.Restore, "Restore from file on startup")

	var storeIntervalSec float64
	flag.Float64Var(&storeIntervalSec, "i", sc.StoreInterval.Seconds(), "Store interval in seconds")

	var dbDsnStr string
	flag.StringVar(&dbDsnStr, "d", "", "DSN for conn to db")

	flag.Parse()

	// Конвертируем в time.Duration
	sc.StoreInterval = time.Duration(storeIntervalSec * float64(time.Second))
	if dbDsnStr != "" {
		conn, err := db.NewPgConnDsn(dbDsnStr)
		if err != nil {
			return err
		}
		if conn != nil {
			sc.DbDsn = *conn
		}
	}

	// Проверяем, что не переданы неизвестные флаги
	if flag.NArg() > 0 {
		flag.Usage()
		return fmt.Errorf("unknown flags or arguments: %v", flag.Args())
	}
	return nil

}

func (sc *ServerConfig) envParse() error {
	tmpCfg := struct {
		Address         *string  `env:"ADDRESS"`
		LogLevel        *string  `env:"LOG_LEVEL"`
		StoreInterval   *float64 `env:"STORE_INTERVAL"`
		FileStoragePath *string  `env:"FILE_STORAGE_PATH"`
		Restore         *bool    `env:"RESTORE"`
		DbDsn           *string  `env:"DATABASE_DSN"`
	}{}

	err := env.Parse(&tmpCfg)
	if err != nil {
		return err
	}

	if tmpCfg.Address != nil {
		sc.Address = *tmpCfg.Address
	}
	if tmpCfg.LogLevel != nil {
		sc.LogLevel = *tmpCfg.LogLevel
	}
	if tmpCfg.StoreInterval != nil {
		sc.StoreInterval = time.Duration(*tmpCfg.StoreInterval * float64(time.Second))
	}
	if tmpCfg.FileStoragePath != nil {
		sc.FileStoragePath = *tmpCfg.FileStoragePath
	}
	if tmpCfg.Restore != nil {
		sc.Restore = *tmpCfg.Restore
	}
	if tmpCfg.DbDsn != nil {
		conn, err := db.NewPgConnDsn(*tmpCfg.DbDsn)
		if err != nil {
			return err
		}
		if conn != nil {
			sc.DbDsn = *conn
		}
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
	if sc.StoreInterval < 0 {
		errs = append(errs, fmt.Errorf("store interval must be non-negative, got: %s", sc.StoreInterval))
	}
	if sc.FileStoragePath == "" {
		errs = append(errs, fmt.Errorf("file storage path must be set, got empty value"))
	}
	return errors.Join(errs...)
}
