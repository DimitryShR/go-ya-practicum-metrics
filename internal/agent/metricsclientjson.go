package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/compress"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

// SendMetricJSON отправляет одну метрику в JSON-формате (использует background context).
func (c *MetricsClient) SendMetricJSON(metric models.Metrics) error {
	return c.SendMetricJSONWithContext(context.Background(), metric)
}

// SendMetricJSONWithContext отправляет одну метрику в JSON-формате на эндпоинт /update.
func (c *MetricsClient) SendMetricJSONWithContext(ctx context.Context, metric models.Metrics) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		return fmt.Errorf("failed to encode metric to JSON: %w", err)
	}
	compressedBody, err := compress.GzipData(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compress request body: %w", err)
	}

	url := fmt.Sprintf("%s/update", c.config.ServerAddress)

	return c.withRetry(ctx, func() error {
		req := c.client.R().SetContext(ctx)
		c.setHashHeader(req, buf.Bytes())
		resp, err := req.
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(compressedBody).
			Post(url)

		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			return NewHTTPStatusError(
				fmt.Errorf("server returned status: %d", resp.StatusCode()),
				resp.StatusCode(),
			)
		}

		return nil
	})
}

// SendAllMetricJSON отправляет массив метрик последовательно в JSON-формате.
func (c *MetricsClient) SendAllMetricJSON(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := c.SendMetricJSON(metric); err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}
	}
	return nil
}

// BatchSendMetricsJSON отправляет массив метрик пакетом в JSON-формате на эндпоинт /updates.
func (c *MetricsClient) BatchSendMetricsJSON(metrics []models.Metrics) error {

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(metrics); err != nil {
		return fmt.Errorf("failed to encode metric to JSON: %w", err)
	}

	compressedBody, err := compress.GzipData(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compress request body: %w", err)
	}

	url := fmt.Sprintf("%s/updates", c.config.ServerAddress)

	return c.withRetry(context.Background(), func() error {
		req := c.client.R()
		c.setHashHeader(req, buf.Bytes())
		resp, err := req.
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(compressedBody).
			Post(url)

		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			return NewHTTPStatusError(fmt.Errorf("server returned status: %d", resp.StatusCode()), resp.StatusCode())
		}

		return nil
	})
}
