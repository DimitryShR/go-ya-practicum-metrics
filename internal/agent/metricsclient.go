package agent

import (
	"fmt"
	"net/http"
	"time"

	config "github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
)

type MetricsClient struct {
	config     *config.AgentConfig
	httpClient *http.Client
}

func NewMetricsClient(cfg *config.AgentConfig) *MetricsClient {
	return &MetricsClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *MetricsClient) GetMetricUrl(metric models.Metrics) (string, error) {
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

func (c *MetricsClient) SendMetric(metric models.Metrics) error {
	url, err := c.GetMetricUrl(metric)
	if err != nil {
		return fmt.Errorf("failed to get metric URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *MetricsClient) SendMetrics(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := c.SendMetric(metric); err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}
	}
	return nil
}
