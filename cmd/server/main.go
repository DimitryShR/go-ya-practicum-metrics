package main

import (
	"net/http"

	handler "github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	middleware "github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	repository "github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	service "github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	storage := repository.NewMemStorage()
	metricService := service.NewMetricService(storage)
	metricHandler := handler.NewMetricHandler(metricService)

	mux := http.NewServeMux()

	updateHandler := middleware.ParseUpdatePath(http.HandlerFunc(metricHandler.UpdateMetricHandler))

	mux.HandleFunc(`/update/`, updateHandler)
	return http.ListenAndServe(`:8080`, mux)
}
