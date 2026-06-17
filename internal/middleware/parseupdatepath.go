package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

// MetricKey — тип ключа для контекста запроса.
type MetricKey string

// Metric — ключ для хранения распаршенной метрики в контексте запроса.
const Metric MetricKey = "metric"

// ParseUpdatePathHandler — middleware для парсинга URL-пути /update/{type}/{name}/{value}.
// Извлекает тип, имя и значение метрики из пути и сохраняет их в контекст запроса.
func ParseUpdatePathHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Middleware logic to parse the update path
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		// get the metric type, name and value from the URL
		remainingPath := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(remainingPath, "/")
		if len(parts) < 4 || parts[0] != "update" {
			http.Error(w, "Invalid path format", http.StatusNotFound)
			return
		}

		metricType := parts[1]
		metricName := parts[2]
		metricValue := parts[3]

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		var metric models.Metrics

		switch models.MetricType(metricType) {
		case models.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid Gauge metric value", http.StatusBadRequest)
				return
			}
			metric = models.Metrics{
				ID:    metricName,
				MType: models.Gauge,
				Value: &value,
			}

		case models.Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid Counter metric value", http.StatusBadRequest)
				return
			}
			metric = models.Metrics{
				ID:    metricName,
				MType: models.Counter,
				Delta: &value,
			}
		default:
			http.Error(w, "Invalid metric type: "+metricType, http.StatusBadRequest)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), Metric, metric))

		next.ServeHTTP(w, r)
	})
}
