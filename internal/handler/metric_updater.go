package handler

import (
	"net/http"

	middleware "github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
	repository "github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
)

type MetricHandler struct {
	storage repository.MemStorage
}

func NewMetricHandler(storage *repository.MemStorage) *MetricHandler {
	return &MetricHandler{storage: *storage}
}

func (h *MetricHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	metric, ok := r.Context().Value(middleware.Metric).(models.Metrics)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.storage.UpdateMetric(metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
