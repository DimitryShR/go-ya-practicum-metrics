package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/compress"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

// метод отправки одной метрики
func (c *MetricsClient) SendMetricJSON(metric models.Metrics) error {

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		return fmt.Errorf("failed to encode metric to JSON: %w", err)
	}

	compressedBody, err := compress.GzipData(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compress request body: %w", err)
	}

	url := fmt.Sprintf("%s/update", c.config.ServerAddress)

	resp, err := c.client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(compressedBody).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode())
	}

	return nil
}

// метод для отправки всех метрик
func (c *MetricsClient) SendAllMetricJSON(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := c.SendMetricJSON(metric); err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}
	}
	return nil
}
