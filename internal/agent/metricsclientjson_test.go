package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

func TestMetricsClient_SendMetricJson(t *testing.T) {
	t.Run("Successful send", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
			}
			var got models.Metrics
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}
			if got.ID != "testMetric" || got.MType != models.Gauge || got.Value == nil || *got.Value != 10.5 {
				t.Fatalf("unexpected metric in body: %+v", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.NewTestAgentConfig(server.URL)
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: func() *float64 { v := 10.5; return &v }(),
		}

		if err := client.SendMetricJson(metric); err != nil {
			t.Errorf("SendMetricJson() error = %v", err)
		}
	})

	t.Run("Server returns error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := config.NewTestAgentConfig(server.URL)
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: func() *float64 { v := 10.5; return &v }(),
		}

		err := client.SendMetricJson(metric)
		if err == nil {
			t.Error("Expected error but got none")
		}
		expectedErr := fmt.Sprintf("server returned status: %d", http.StatusInternalServerError)
		if err == nil || err.Error() != expectedErr {
			t.Errorf("Expected server error, got %v", err)
		}
	})

	t.Run("Network error", func(t *testing.T) {

		cfg := config.NewTestAgentConfig("http://invalid-server:9999")
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		// Проверяем, что httpClient существует перед установкой timeout
		if client.client == nil {
			t.Fatal("httpClient is nil")
		}
		// Уменьшаем timeout для быстрого падения теста
		client.client.SetTimeout(100 * time.Millisecond)
		// Отключаем ретраи для данного теста
		client.client.SetRetryCount(0)
		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: func() *float64 { v := 10.5; return &v }(),
		}

		err := client.SendMetricJson(metric)
		if err == nil {
			t.Error("Expected network error but got none")
		}
	})
}

func TestMetricsClient_SendAllMetricJson(t *testing.T) {
	t.Run("Successful send multiple metrics", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.NewTestAgentConfig(server.URL)
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		metrics := []models.Metrics{
			{
				MType: models.Gauge,
				ID:    "metric1",
				Value: func() *float64 { v := 1.0; return &v }(),
			},
			{
				MType: models.Counter,
				ID:    "metric2",
				Delta: func() *int64 { v := int64(2); return &v }(),
			},
			{
				MType: models.Gauge,
				ID:    "metric3",
				Value: func() *float64 { v := 3.0; return &v }(),
			},
		}

		if err := client.SendAllMetricJson(metrics); err != nil {
			t.Errorf("SendAllMetricJson() error = %v", err)
		}

		if requestCount != 3 {
			t.Errorf("Expected 3 requests, got %d", requestCount)
		}
	})

	t.Run("Failed to send one metric stops processing", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount == 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.NewTestAgentConfig(server.URL)
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		metrics := []models.Metrics{
			{
				MType: models.Gauge,
				ID:    "metric1",
				Value: func() *float64 { v := 1.0; return &v }(),
			},
			{
				MType: models.Gauge,
				ID:    "metric2",
				Value: func() *float64 { v := 2.0; return &v }(),
			},
			{
				MType: models.Gauge,
				ID:    "metric3",
				Value: func() *float64 { v := 3.0; return &v }(),
			},
		}

		err := client.SendAllMetricJson(metrics)
		if err == nil {
			t.Error("Expected error but got none")
		}

		if requestCount != 2 {
			t.Errorf("Expected only 2 requests before error, got %d", requestCount)
		}
	})

	t.Run("Empty metrics slice", func(t *testing.T) {

		cfg := config.NewTestAgentConfig("http://localhost:8080")
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		if err := client.SendAllMetricJson([]models.Metrics{}); err != nil {
			t.Errorf("SendAllMetricJson with empty slice should not error, got %v", err)
		}
	})
}
