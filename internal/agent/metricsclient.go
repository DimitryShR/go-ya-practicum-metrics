package agent

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/sign"
	"github.com/go-resty/resty/v2"
)

// MetricsClient — HTTP-клиент для отправки метрик на сервер.
// Использует resty для выполнения запросов.
type MetricsClient struct {
	config *config.AgentConfig
	client *resty.Client
	signer *sign.Signer
}

// NewMetricsClient создаёт новый MetricsClient с базовым URL сервера.
func NewMetricsClient(cfg *config.AgentConfig) *MetricsClient {
	restyClient := resty.New()
	restyClient.SetTimeout(5 * time.Second)

	var signer *sign.Signer
	if cfg.SignKey != "" {
		signer = sign.NewSigner(cfg.SignKey)
	}

	return &MetricsClient{
		config: cfg,
		client: restyClient,
		signer: signer,
	}
}

// joinURL объединяет базовый URL с сегментами пути, корректно обрабатывая двойной слэш схемы.
func joinURL(base string, parts ...string) string {
	return strings.TrimRight(base, "/") + "/" + path.Join(parts...)
}

// setHashHeader устанавливает заголовок HashSHA256, если настроен signer.
func (c *MetricsClient) setHashHeader(r *resty.Request, body []byte) *resty.Request {
	if c.signer == nil {
		return r
	}
	return r.SetHeader("HashSHA256", c.signer.Sign(body))
}

// getMetricURL формирует URL для отправки метрики через URL-путь.
func (c *MetricsClient) getMetricURL(metric models.Metrics) (string, error) {
	var url string

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return "", fmt.Errorf("gauge metric value is nil")
		}
		url = joinURL(c.config.ServerAddress, "update",
			string(metric.MType), metric.ID, fmt.Sprintf("%v", *metric.Value))

	case models.Counter:
		if metric.Delta == nil {
			return "", fmt.Errorf("counter metric delta is nil")
		}
		url = joinURL(c.config.ServerAddress, "update",
			string(metric.MType), metric.ID, fmt.Sprintf("%v", *metric.Delta))

	default:
		return "", fmt.Errorf("unknown metric type: %s", metric.MType)
	}
	return url, nil
}

// SendMetric отправляет одну метрику через URL-путь (использует background context).
func (c *MetricsClient) SendMetric(metric models.Metrics) error {
	return c.SendMetricWithContext(context.Background(), metric)
}

// SendMetricWithContext отправляет одну метрику через URL-путь /update/{type}/{name}/{value}.
func (c *MetricsClient) SendMetricWithContext(ctx context.Context, metric models.Metrics) error {
	if ctx == nil {
		ctx = context.Background()
	}
	url, err := c.getMetricURL(metric)
	if err != nil {
		return fmt.Errorf("failed to get metric URL: %w", err)
	}

	return c.withRetry(ctx, func() error {
		req := c.client.R().SetContext(ctx)
		resp, err := req.SetHeader("Content-Type", "text/plain").Post(url)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("server returned status: %d", resp.StatusCode())
		}

		return nil
	})
}

// SendMetrics отправляет массив метрик последовательно.
func (c *MetricsClient) SendMetrics(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := c.SendMetric(metric); err != nil {
			return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
		}
	}
	return nil
}

// GetMetric получает значение метрики через URL-путь (использует background context).
func (c *MetricsClient) GetMetric(metricType models.MetricType, metricName string) (string, error) {
	return c.GetMetricWithContext(context.Background(), metricType, metricName)
}

// GetMetricWithContext получает значение метрики через URL-путь /value/{type}/{name}.
func (c *MetricsClient) GetMetricWithContext(
	ctx context.Context,
	metricType models.MetricType,
	metricName string,
) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	url := joinURL(c.config.ServerAddress, "value", string(metricType), metricName)

	var body string
	err := c.withRetry(ctx, func() error {
		req := c.client.R().SetContext(ctx)
		resp, err := req.Get(url)
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
