package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	config "github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
)

var defaultCfg *config.AgentConfig = config.NewDefaultAgentConfig()

func TestNewMetricsClient(t *testing.T) {
	client := NewMetricsClient(defaultCfg)

	if client == nil {
		t.Fatal("Expected MetricsClient to be created, got nil")
	}

	if client.config != defaultCfg {
		t.Error("Expected config to be set")
	}
}

func TestMetricsClient_getMetricURL(t *testing.T) {
	tests := []struct {
		name     string
		metric   models.Metrics
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid gauge metric",
			metric: models.Metrics{
				MType: models.Gauge,
				ID:    "testGauge",
				Value: func() *float64 { v := 123.45; return &v }(),
			},
			expected: "http://localhost:8080/update/gauge/testGauge/123.45",
			wantErr:  false,
		},
		{
			name: "Gauge metric with nil value",
			metric: models.Metrics{
				MType: models.Gauge,
				ID:    "testGauge",
				Value: nil,
			},
			wantErr: true,
			errMsg:  "gauge metric value is nil",
		},
		{
			name: "Valid counter metric",
			metric: models.Metrics{
				MType: models.Counter,
				ID:    "testCounter",
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
			expected: "http://localhost:8080/update/counter/testCounter/100",
			wantErr:  false,
		},
		{
			name: "Counter metric with nil delta",
			metric: models.Metrics{
				MType: models.Counter,
				ID:    "testCounter",
				Delta: nil,
			},
			wantErr: true,
			errMsg:  "counter metric delta is nil",
		},
		{
			name: "Unknown metric type",
			metric: models.Metrics{
				MType: "unknown",
				ID:    "testUnknown",
			},
			wantErr: true,
			errMsg:  "unknown metric type: unknown",
		},
	}

	client := NewMetricsClient(defaultCfg)
	if client == nil {
		t.Fatal("Failed to create MetricsClient")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.getMetricURL(tt.metric)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("GetMetricURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMetricsClient_SendMetric(t *testing.T) {
	t.Run("Successful send", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "text/plain" {
				t.Errorf("Expected Content-Type: text/plain, got %s", r.Header.Get("Content-Type"))
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.NewCustomServerAddressAgentConfig(server.URL)
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: func() *float64 { v := 10.5; return &v }(),
		}

		if err := client.SendMetric(metric); err != nil {
			t.Errorf("SendMetric() error = %v", err)
		}
	})

	t.Run("Server returns error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := config.NewCustomServerAddressAgentConfig(server.URL)
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: func() *float64 { v := 10.5; return &v }(),
		}

		err := client.SendMetric(metric)
		if err == nil {
			t.Error("Expected error but got none")
		}
		expectedErr := fmt.Sprintf("server returned status: %d", http.StatusInternalServerError)
		if err == nil || err.Error() != expectedErr {
			t.Errorf("Expected server error, got %v", err)
		}
	})

	t.Run("Network error", func(t *testing.T) {

		cfg := config.NewCustomServerAddressAgentConfig("http://invalid-server:9999")
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

		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: func() *float64 { v := 10.5; return &v }(),
		}

		err := client.SendMetric(metric)
		if err == nil {
			t.Error("Expected network error but got none")
		}
	})

	t.Run("Invalid URL from getMetricURL", func(t *testing.T) {

		cfg := config.NewCustomServerAddressAgentConfig("http://localhost:8080")
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}
		// Метрика с nil значением вызовет ошибку в getMetricURL
		metric := models.Metrics{
			MType: models.Gauge,
			ID:    "testMetric",
			Value: nil,
		}

		err := client.SendMetric(metric)
		if err == nil {
			t.Error("Expected error from getMetricURL but got none")
		}
		expectedErr := "failed to get metric URL: gauge metric value is nil"
		if err == nil || err.Error() != expectedErr {
			t.Errorf("Expected getMetricURL error, got %v", err)
		}
	})
}

func TestMetricsClient_SendMetrics(t *testing.T) {
	t.Run("Successful send multiple metrics", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.NewCustomServerAddressAgentConfig(server.URL)
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

		if err := client.SendMetrics(metrics); err != nil {
			t.Errorf("SendMetrics() error = %v", err)
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

		cfg := config.NewCustomServerAddressAgentConfig(server.URL)
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

		err := client.SendMetrics(metrics)
		if err == nil {
			t.Error("Expected error but got none")
		}

		if requestCount != 2 {
			t.Errorf("Expected only 2 requests before error, got %d", requestCount)
		}
	})

	t.Run("Empty metrics slice", func(t *testing.T) {

		cfg := config.NewCustomServerAddressAgentConfig("http://localhost:8080")
		client := NewMetricsClient(cfg)
		if client == nil {
			t.Fatal("Failed to create MetricsClient")
		}

		if err := client.SendMetrics([]models.Metrics{}); err != nil {
			t.Errorf("SendMetrics with empty slice should not error, got %v", err)
		}
	})
}
