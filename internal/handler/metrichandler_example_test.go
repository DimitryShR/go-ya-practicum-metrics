package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/audit"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
)

// ExampleMetricHandler_UpdateMetricHandler демонстрирует обновление метрики через URL-путь.
// В реальном приложении middleware ParseUpdatePathHandler парсит URL и кладёт метрику в контекст.
func ExampleMetricHandler_UpdateMetricHandler() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricService(repo)
	pub := audit.NewAuditPublisher()
	h := NewMetricHandler(svc, pub)

	// Пример: POST /update/gauge/test_metric/42.5
	// Middleware ParseUpdatePathHandler должен быть вызван перед этим обработчиком
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/42.5", nil)
	w := httptest.NewRecorder()

	// В реальном коде middleware кладёт метрику в контекст:
	// ctx := context.WithValue(req.Context(), middleware.Metric, metric)
	// req = req.WithContext(ctx)

	h.UpdateMetricHandler(w, req)

	fmt.Println("Status:", w.Code)
}

// ExampleMetricHandler_GetMetricValue демонстрирует получение значения метрики через URL-путь.
func ExampleMetricHandler_GetMetricValue() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricService(repo)
	pub := audit.NewAuditPublisher()
	h := NewMetricHandler(svc, pub)

	// Сначала обновляем метрику через сервис
	_ = svc.UpdateMetric(context.TODO(), models.Metrics{
		ID:    "test_counter",
		MType: models.Counter,
		Delta: func() *int64 { v := int64(10); return &v }(),
	})

	// GET /value/counter/test_counter
	req := httptest.NewRequest(http.MethodGet, "/value/counter/test_counter", nil)
	w := httptest.NewRecorder()

	h.GetMetricValue(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())
}

// ExampleMetricHandler_GetAllMetrics демонстрирует получение всех метрик в HTML.
func ExampleMetricHandler_GetAllMetrics() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricService(repo)
	pub := audit.NewAuditPublisher()
	h := NewMetricHandler(svc, pub)

	_ = svc.UpdateMetric(context.TODO(), models.Metrics{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: func() *float64 { v := 3.14; return &v }(),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.GetAllMetrics(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))
	fmt.Println("Body contains test_gauge:", strings.Contains(w.Body.String(), "test_gauge"))
}
