package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"database/sql"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type storageMode string

const (
	modeMemory   storageMode = "memory"
	modeFile     storageMode = "file"
	modePostgres storageMode = "postgres"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var storage service.Storage
	var db *sql.DB

	cfg := config.NewServerConfig()

	fmt.Printf("Starting server on %s\n", cfg.Address)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return err
	}
	defer logger.Log.Sync()

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	mode := getMode(cfg)
	logger.Log.Info("Storage mode selected", zap.String("mode", string(mode)))

	switch mode {
	case modePostgres:
		pgStorage, err := initPostgresStorage(cfg, &db)
		if err != nil {
			return fmt.Errorf("initialize PostgreSQL storage: %w", err)
		}
		if db != nil {
			defer db.Close()
		}
		storage = pgStorage
	case modeFile:
		fileStorage, err := initFileStorage(cfg)
		if err != nil {
			return fmt.Errorf("initialize file storage: %w", err)
		}
		storage = fileStorage
	case modeMemory:
		storage = repository.NewMemStorage()
	}

	metricService := service.NewMetricService(storage)
	metricHandler := handler.NewMetricHandler(metricService)

	pingHandler := handler.NewPingHandler(db)
	router := newRouter(metricHandler, pingHandler)

	return http.ListenAndServe(cfg.Address, router)
}

func newRouter(metricHandler *handler.MetricHandler, pingHandler *handler.PingHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.LogRequest, middleware.GzipMiddleware)

	r.Post("/update", metricHandler.UpdateMetricHandlerJSON)
	r.With(middleware.ParseUpdatePathHandler).Post("/update/*", metricHandler.UpdateMetricHandler)

	r.Post("/value", metricHandler.GetMetricValueJSON)
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetricValue)
	r.Get("/", metricHandler.GetAllMetrics)
	r.Get("/ping", pingHandler.Ping)

	return r
}

func getMode(cfg *config.ServerConfig) storageMode {
	dbDSN := cfg.DBDsn.GetKeywordDSN()
	if dbDSN != "" {
		return modePostgres
	}
	if cfg.FileStoragePath != "" {
		return modeFile
	}
	return modeMemory
}

func initPostgresStorage(cfg *config.ServerConfig, db **sql.DB) (service.Storage, error) {
	migrateURL := cfg.DBMigrateDsn.GetURL()
	if migrateURL == "" {
		logger.Log.Info("DB migrate DSN not provided, try using main DB DSN for migrations")
		migrateURL = cfg.DBDsn.GetURL()
	}
	if err := runMigrations(migrateURL); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	logger.Log.Info("Database migrations completed successfully")

	dbDSN := cfg.DBDsn.GetKeywordDSN()
	newDB, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := newDB.PingContext(ctx); err != nil {
		_ = newDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	*db = newDB
	return repository.NewPgStorage(newDB), nil
}

func runMigrations(migrateDSN string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory for migrations: %w", err)
	}
	sourceURL := "file://" + filepath.Join(wd, "migrations")

	m, err := migrate.New(sourceURL, migrateDSN)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func initFileStorage(cfg *config.ServerConfig) (service.Storage, error) {
	memStorage := repository.NewMemStorage()

	if cfg.Restore {
		if ok, err := memStorage.LoadFromFile(cfg.FileStoragePath); err != nil {
			return nil, fmt.Errorf("restore metrics from file: %w", err)
		} else if ok {
			logger.Log.Info("Metrics restored from file", zap.String("file", cfg.FileStoragePath))
		}
	}

	switch {
	case cfg.StoreInterval < 0:
		return nil, fmt.Errorf("store interval must be non-negative, got %s", cfg.StoreInterval)
	case cfg.StoreInterval == 0:
		memStorage.EnableSyncSave(cfg.FileStoragePath)
		logger.Log.Info("Enabled synchronous metrics persistence", zap.String("file", cfg.FileStoragePath))
	default:
		go func(path string, interval time.Duration) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for range ticker.C {
				if err := memStorage.SaveToFile(path); err != nil {
					logger.Log.Error("Cannot save metrics to file", zap.Error(err), zap.String("file", path))
				}
			}
		}(cfg.FileStoragePath, cfg.StoreInterval)
		logger.Log.Info(
			"Enabled periodic metrics persistence",
			zap.Duration("interval", cfg.StoreInterval),
			zap.String("file", cfg.FileStoragePath),
		)
	}

	return memStorage, nil
}
