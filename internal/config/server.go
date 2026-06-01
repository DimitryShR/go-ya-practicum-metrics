package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	DBDsn           db.PgConn     // `env:"DATABASE_DSN"`
	DBMigrateDsn    db.PgConn     // `env:"DATABASE_MIGRATE_DSN"`
	SignKey         string        // `env:"KEY"`
}

// Создаем новый экземпляр конфигурации сервера, загружая значения конфигурации
// Приоритет загрузки: переменные окружения > флаги > значения по умолчанию
func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Address:         ":8080",
		LogLevel:        "info",
		StoreInterval:   300 * time.Second,
		FileStoragePath: "",
		Restore:         true,
		DBDsn:           db.PgConn{},
		DBMigrateDsn:    db.PgConn{},
		SignKey:         "",
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
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	// Подавляем вывод
	fs.SetOutput(io.Discard)

	return sc.parseFlagSet(fs, os.Args[1:])
}

func (sc *ServerConfig) parseFlagSet(fs *flag.FlagSet, args []string) error {
	fs.StringVar(&sc.Address, "a", sc.Address, "Server address")
	fs.StringVar(&sc.LogLevel, "loglvl", sc.LogLevel, "Log level")
	fs.StringVar(&sc.FileStoragePath, "f", sc.FileStoragePath, "File storage path")
	fs.BoolVar(&sc.Restore, "r", sc.Restore, "Restore from file on startup")

	var storeIntervalSec float64
	fs.Float64Var(&storeIntervalSec, "i", sc.StoreInterval.Seconds(), "Store interval in seconds")

	// Подключение к БД через флаг -d для основного подключения
	var dbDsnStr string
	fs.StringVar(&dbDsnStr, "d", "", "DSN for conn to db")

	// Подключение к БД для миграции через флаг -m
	var dbMigrateDsnStr string
	fs.StringVar(&dbMigrateDsnStr, "m", "", "DSN for migrate to db")

	fs.StringVar(&sc.SignKey, "k", sc.SignKey, "Key for sign data")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Конвертируем в time.Duration
	sc.StoreInterval = time.Duration(storeIntervalSec * float64(time.Second))
	if dbDsnStr != "" {
		conn, err := db.NewPgConnDsn(dbDsnStr)
		if err != nil {
			return err
		}
		if conn != nil {
			sc.DBDsn = *conn
		}
	}
	if dbMigrateDsnStr != "" {
		conn, err := db.NewPgConnDsn(dbMigrateDsnStr)
		if err != nil {
			return err
		}
		if conn != nil {
			sc.DBMigrateDsn = *conn
		}
	}

	// Проверяем, что не переданы неизвестные флаги
	if fs.NArg() > 0 {
		return fmt.Errorf("unknown flags or arguments: %v", fs.Args())
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
		DBDsn           *string  `env:"DATABASE_DSN"`
		DBMigrateDsn    *string  `env:"DATABASE_MIGRATE_DSN"`
		SignKey         *string  `env:"KEY"`
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
	if tmpCfg.DBDsn != nil {
		conn, err := db.NewPgConnDsn(*tmpCfg.DBDsn)
		if err != nil {
			return err
		}
		if conn != nil {
			sc.DBDsn = *conn
		}
	}
	if tmpCfg.DBMigrateDsn != nil {
		conn, err := db.NewPgConnDsn(*tmpCfg.DBMigrateDsn)
		if err != nil {
			return err
		}
		if conn != nil {
			sc.DBMigrateDsn = *conn
		}
	}
	if tmpCfg.SignKey != nil {
		sc.SignKey = *tmpCfg.SignKey
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
	// if sc.FileStoragePath == "" {
	// 	errs = append(errs, fmt.Errorf("file storage path must be set, got empty value"))
	// }
	return errors.Join(errs...)
}
