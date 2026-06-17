// Package handler предоставляет HTTP-обработчики.
package handler

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/audit"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"go.uber.org/zap"
)

// metricsTemplate создаётся один раз при инициализации приложения.
var metricsTemplate = func() *template.Template {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
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
	t, err := template.New("metrics").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		panic("failed to parse metrics template: " + err.Error())
	}
	return t
}()

// MetricHandler — HTTP-обработчик для работы с метриками.
type MetricHandler struct {
	service   service.MetricService
	publisher audit.Publisher
}

// NewMetricHandler создаёт новый MetricHandler с указанным сервисом метрик и издателем аудит-событий.
func NewMetricHandler(service service.MetricService, publisher audit.Publisher) *MetricHandler {
	return &MetricHandler{
		service:   service,
		publisher: publisher,
	}
}

// extractIPAddress извлекает IP-адрес из RemoteAddr запроса.
func extractIPAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// UpdateMetricHandler — обработчик POST /update/{type}/{name}/{value}.
// Обновляет метрику через URL-параметры. После успешного обновления отправляет аудит-событие.
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
	if mh.publisher != nil {
		mh.publisher.Notify(audit.AuditEvent{
			Timestamp: time.Now().Unix(),
			Metrics:   []string{metric.ID},
			IPAddress: extractIPAddress(r),
		})
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetMetricValue — обработчик GET /value/{type}/{name}.
// Возвращает значение метрики указанного типа и имени.
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
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", value)

	case models.Counter:
		value, err := mh.service.GetCounter(r.Context(), metricName)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", value)

	default:
		http.Error(w, "Invalid metric type", http.StatusUnprocessableEntity)
	}
}

// GetAllMetrics — обработчик GET /.
// Возвращает HTML-страницу со всеми сохранёнными метриками (gauge и counter).
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

	data := struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := metricsTemplate.Execute(w, data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}
