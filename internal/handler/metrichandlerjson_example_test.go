package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/audit"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
)

// ExampleMetricHandler_UpdateMetricHandlerJSON демонстрирует обновление метрики через JSON API.
func ExampleMetricHandler_UpdateMetricHandlerJSON() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricService(repo)
	pub := audit.NewAuditPublisher()
	h := NewMetricHandler(svc, pub)

	metric := models.Metrics{
		ID:    "test_gauge_json",
		MType: models.Gauge,
		Value: func() *float64 { v := 123.45; return &v }(),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateMetricHandlerJSON(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())
}

// ExampleMetricHandler_GetMetricValueJSON демонстрирует получение значения метрики через JSON API.
func ExampleMetricHandler_GetMetricValueJSON() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricService(repo)
	pub := audit.NewAuditPublisher()
	h := NewMetricHandler(svc, pub)

	// Сначала обновляем метрику
	_ = svc.UpdateMetric(context.TODO(), models.Metrics{
		ID:    "test_counter_json",
		MType: models.Counter,
		Delta: func() *int64 { v := int64(42); return &v }(),
	})

	requestMetric := models.Metrics{
		ID:    "test_counter_json",
		MType: models.Counter,
	}
	body, _ := json.Marshal(requestMetric)

	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.GetMetricValueJSON(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())
}

// ExampleMetricHandler_UpdateMetricsHandlerJSON демонстрирует пакетное обновление метрик через JSON API.
func ExampleMetricHandler_UpdateMetricsHandlerJSON() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricService(repo)
	pub := audit.NewAuditPublisher()
	h := NewMetricHandler(svc, pub)

	metrics := []models.Metrics{
		{
			ID:    "batch_gauge",
			MType: models.Gauge,
			Value: func() *float64 { v := 1.1; return &v }(),
		},
		{
			ID:    "batch_counter",
			MType: models.Counter,
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateMetricsHandlerJSON(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())
}
