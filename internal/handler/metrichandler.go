package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"go.uber.org/zap"
)

type MetricHandler struct {
	service service.MetricService
}

func NewMetricHandler(service service.MetricService) *MetricHandler {
	return &MetricHandler{service: service}
}

// UpdateMetric - обработчик POST /update/<type>/<name>/<value>
func (mh *MetricHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Info("got request with bad method", zap.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metric, ok := r.Context().Value(middleware.Metric).(models.Metrics)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := mh.service.UpdateMetric(r.Context(), metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetMetricValue - обработчик GET /value/<type>/<name>
func (mh *MetricHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 3 || parts[0] != "value" {
		http.Error(w, "Invalid path format", http.StatusNotFound)
		return
	}

	metricType := parts[1]
	metricName := parts[2]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	switch models.MetricType(metricType) {
	case models.Gauge:
		value, err := mh.service.GetGauge(r.Context(), metricName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			logger.Log.Error("get gauge failed", zap.Error(err), zap.String("name", metricName))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", value)

	case models.Counter:
		value, err := mh.service.GetCounter(r.Context(), metricName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			logger.Log.Error("get counter failed", zap.Error(err), zap.String("name", metricName))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", value)

	default:
		http.Error(w, "Invalid metric type", http.StatusUnprocessableEntity)
	}
}

// GetAllMetrics - обработчик GET / (HTML страница со всеми метриками)
func (mh *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	gauges, err := mh.service.GetAllGauges(r.Context())
	if err != nil {
		logger.Log.Error("get all gauges failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	counters, err := mh.service.GetAllCounters(r.Context())
	if err != nil {
		logger.Log.Error("get all counters failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// HTML шаблон
	tmpl := `
<!DOCTYPE html>
<html>
<body>
    <h1>Metrics</h1>

    {{range $name, $value := .Gauges}}
    <div>{{$name}}: {{printf "%.2f" $value}}</div>
    {{end}}

    {{range $name, $value := .Counters}}
    <div>{{$name}}: {{$value}}</div>
    {{end}}
</body>
</html>
`

	// Функции для шаблона
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	t, err := template.New("metrics").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}
