package main

import (
	"fmt"
	"net/http"

	config "github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	handler "github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	middleware "github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	repository "github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	service "github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.NewServerConfig()
	fmt.Printf("Starting server on %s\n", cfg.Address)

	storage := repository.NewMemStorage()
	metricService := service.NewMetricService(storage)
	metricHandler := handler.NewMetricHandler(metricService)

	r := chi.NewRouter()

	r.With(middleware.ParseUpdatePathHandler).Post("/update/*", metricHandler.UpdateMetricHandler)

	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetricValue)
	r.Get("/", metricHandler.GetAllMetrics)

	return http.ListenAndServe(cfg.Address, r)
}
