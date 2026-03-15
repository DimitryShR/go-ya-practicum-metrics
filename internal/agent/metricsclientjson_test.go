package agent

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

func TestMetricsClient_SendMetricJSON(t *testing.T) {
	t.Run("Successful send", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
			}
			if r.Header.Get("Content-Encoding") != "gzip" {
				t.Errorf("Expected Content-Encoding: gzip, got %s", r.Header.Get("Content-Encoding"))
			}

			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Fatalf("failed to create gzip reader: %v", err)
			}
			defer gr.Close()

			var got models.Metrics
			if err := json.NewDecoder(gr).Decode(&got); err != nil {
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

		if err := client.SendMetricJSON(metric); err != nil {
			t.Errorf("SendMetricJSON() error = %v", err)
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

		err := client.SendMetricJSON(metric)
		if err == nil {
			t.Error("Expected error but got none")
		}
		var statusErr *HTTPStatusError
		if !errors.As(err, &statusErr) {
			t.Fatalf("Expected HTTPStatusError, got %T", err)
		}
		if statusErr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, statusErr.Code)
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

		err := client.SendMetricJSON(metric)
		if err == nil {
			t.Error("Expected network error but got none")
		}
	})
}

func TestMetricsClient_SendAllMetricJSON(t *testing.T) {
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

		if err := client.SendAllMetricJSON(metrics); err != nil {
			t.Errorf("SendAllMetricJSON() error = %v", err)
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

		err := client.SendAllMetricJSON(metrics)
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

		if err := client.SendAllMetricJSON([]models.Metrics{}); err != nil {
			t.Errorf("SendAllMetricJSON with empty slice should not error, got %v", err)
		}
	})
}

func TestMetricsClient_BatchSendMetricsJSON(t *testing.T) {
	t.Run("Successful send batch metrics", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/updates" {
				t.Errorf("Expected path /updates, got %s", r.URL.Path)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
			}
			if r.Header.Get("Content-Encoding") != "gzip" {
				t.Errorf("Expected Content-Encoding: gzip, got %s", r.Header.Get("Content-Encoding"))
			}

			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Fatalf("failed to create gzip reader: %v", err)
			}
			defer gr.Close()

			var got []models.Metrics
			if err := json.NewDecoder(gr).Decode(&got); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}

			if len(got) != 2 {
				t.Fatalf("unexpected metrics length: %d", len(got))
			}

			if got[0].ID != "metric1" || got[0].MType != models.Gauge || got[0].Value == nil || *got[0].Value != 1.5 {
				t.Fatalf("unexpected first metric in body: %+v", got[0])
			}
			if got[1].ID != "metric2" || got[1].MType != models.Counter || got[1].Delta == nil || *got[1].Delta != 10 {
				t.Fatalf("unexpected second metric in body: %+v", got[1])
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
				Value: func() *float64 { v := 1.5; return &v }(),
			},
			{
				MType: models.Counter,
				ID:    "metric2",
				Delta: func() *int64 { v := int64(10); return &v }(),
			},
		}

		if err := client.BatchSendMetricsJSON(metrics); err != nil {
			t.Errorf("BatchSendMetricsJSON() error = %v", err)
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

		metrics := []models.Metrics{
			{
				MType: models.Gauge,
				ID:    "metric1",
				Value: func() *float64 { v := 1.5; return &v }(),
			},
		}

		err := client.BatchSendMetricsJSON(metrics)
		if err == nil {
			t.Error("Expected error but got none")
		}
		var statusErr *HTTPStatusError
		if !errors.As(err, &statusErr) {
			t.Fatalf("Expected HTTPStatusError, got %T", err)
		}
		if statusErr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, statusErr.Code)
		}
	})

	t.Run("Network error", func(t *testing.T) {
		cfg := config.NewTestAgentConfig("http://invalid-server:9999")
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		if client.client == nil {
			t.Fatal("httpClient is nil")
		}
		client.client.SetTimeout(100 * time.Millisecond)
		client.client.SetRetryCount(0)

		metrics := []models.Metrics{
			{
				MType: models.Gauge,
				ID:    "metric1",
				Value: func() *float64 { v := 1.5; return &v }(),
			},
		}

		err := client.BatchSendMetricsJSON(metrics)
		if err == nil {
			t.Error("Expected network error but got none")
		}
	})
}
