package main

import (
	"fmt"
	"net/http"
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
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.NewServerConfig()
	fmt.Printf("Starting server on %s\n", cfg.Address)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return err
	}
	defer logger.Log.Sync()

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	// Инициализируем подключение к БД
	var db *sql.DB
	if dsn := cfg.DbDsn.GetDsn(); dsn != "" {
		var err error
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
	}
	pingHandler := handler.NewPingHandler(db)

	storage := repository.NewMemStorage()
	if cfg.Restore {
		if ok, err := storage.LoadFromFile(cfg.FileStoragePath); err != nil {
			return fmt.Errorf("restore metrics from file: %w", err)
		} else if ok {
			logger.Log.Info("Metrics restored from file", zap.String("file", cfg.FileStoragePath))
		}
	}

	switch {
	case cfg.StoreInterval < 0:
		return fmt.Errorf("store interval must be non-negative, got %s", cfg.StoreInterval)
	case cfg.StoreInterval == 0:
		storage.EnableSyncSave(cfg.FileStoragePath)
		logger.Log.Info("Enabled synchronous metrics persistence", zap.String("file", cfg.FileStoragePath))
	default:
		go func(path string, interval time.Duration) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for range ticker.C {
				if err := storage.SaveToFile(path); err != nil {
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

	metricService := service.NewMetricService(storage)
	metricHandler := handler.NewMetricHandler(metricService)

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
