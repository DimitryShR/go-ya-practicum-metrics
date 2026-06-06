package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/audit"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
)

// benchService создаёт реальный MetricService для бенчмарков
func benchService() service.MetricService {
	storage := repository.NewMemStorage()
	return service.NewMetricService(storage)
}

func BenchmarkGetAllMetrics(b *testing.B) {
	metricService := benchService()
	publisher := audit.NewAuditPublisher()
	defer publisher.Shutdown()
	h := handler.NewMetricHandler(metricService, publisher)

	// Заполняем хранилище
	for i := 0; i < 30; i++ {
		val := float64(i) * 1.5
		metric := models.Metrics{
			ID:    "gauge_" + string(rune('A'+i%26)),
			MType: models.Gauge,
			Value: &val,
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.UpdateMetricHandlerJSON(rr, req)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		h.GetAllMetrics(rr, req)
	}
}
