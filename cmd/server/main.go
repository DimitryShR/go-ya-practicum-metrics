package main

import (
	"fmt"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
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

	storage := repository.NewMemStorage()
	metricService := service.NewMetricService(storage)
	metricHandler := handler.NewMetricHandler(metricService)

	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)

	r.Post("/update", metricHandler.UpdateMetricHandlerJSON)
	r.With(middleware.ParseUpdatePathHandler).Post("/update/*", metricHandler.UpdateMetricHandler)

	r.Post("/value", metricHandler.GetMetricValueJSON)
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetricValue)
	r.Get("/", metricHandler.GetAllMetrics)

	return http.ListenAndServe(cfg.Address, middleware.LogRequest(middleware.GzipMiddleware(r)))
}
