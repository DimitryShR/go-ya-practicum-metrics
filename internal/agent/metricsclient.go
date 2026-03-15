package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/go-resty/resty/v2"
)

type MetricsClient struct {
	config *config.AgentConfig
	client *resty.Client
}

func NewMetricsClient(cfg *config.AgentConfig) *MetricsClient {
	restyClient := resty.New()
	restyClient.SetTimeout(5 * time.Second)
	return &MetricsClient{
		config: cfg,
		client: restyClient,
	}
}

// Вспомогательный метод извлечения метрик из URL
func (c *MetricsClient) getMetricURL(metric models.Metrics) (string, error) {
	var url string

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return "", fmt.Errorf("gauge metric value is nil")
		}
		url = fmt.Sprintf("%s/update/%s/%s/%v",
			c.config.ServerAddress, metric.MType, metric.ID, *metric.Value)

	case models.Counter:
		if metric.Delta == nil {
			return "", fmt.Errorf("counter metric delta is nil")
		}
		url = fmt.Sprintf("%s/update/%s/%s/%v",
			c.config.ServerAddress, metric.MType, metric.ID, *metric.Delta)

	default:
		return "", fmt.Errorf("unknown metric type: %s", metric.MType)
	}
	return url, nil
}

// метод отправки одной метрики
func (c *MetricsClient) SendMetric(metric models.Metrics) error {
	url, err := c.getMetricURL(metric)
	if err != nil {
		return fmt.Errorf("failed to get metric URL: %w", err)
	}

	return c.withRetry(func() error {
		resp, err := c.client.R().SetHeader("Content-Type", "text/plain").Post(url)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("server returned status: %d", resp.StatusCode())
		}

		return nil
	})
}

// метод для отправки всех метрик
func (c *MetricsClient) SendMetrics(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := c.SendMetric(metric); err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}
	}
	return nil
}

// Метод получения значения метрики
func (c *MetricsClient) GetMetric(metricType models.MetricType, metricName string) (string, error) {
	url := fmt.Sprintf("%s/value/%s/%s", c.config.ServerAddress, metricType, metricName)

	var body string
	err := c.withRetry(func() error {
		resp, err := c.client.R().Get(url)
		if err != nil {
			return fmt.Errorf("failed to get metric: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("server returned status: %d", resp.StatusCode())
		}
		body = string(resp.Body())
		return nil
	})
	if err != nil {
		return "", err
	}
	return body, nil
}
