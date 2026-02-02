package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
)

type MetricKey string

const Metric MetricKey = "metric"

// Адаптер Middleware для разбора пути /update/<type>/<name>/<value>
func ParseUpdatePathHandler(next http.Handler) http.Handler {
	return ParseUpdatePath(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Middleware функция для разбора пути /update/<type>/<name>/<value>ы
func ParseUpdatePath(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Middleware logic to parse the update path
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		// get the metric type, name and value from the URL
		remainingPath := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(remainingPath, "/")
		if parts[0] != "update" || len(parts) < 4 {
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
