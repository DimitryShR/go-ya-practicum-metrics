package handler

import (
	"net/http"

	middleware "github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
	service "github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
)

type MetricHandler struct {
	service service.MetricService
}

func NewMetricHandler(service service.MetricService) *MetricHandler {
	return &MetricHandler{service: service}
}

func (mh *MetricHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	metric, ok := r.Context().Value(middleware.Metric).(models.Metrics)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := mh.service.UpdateMetric(metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
