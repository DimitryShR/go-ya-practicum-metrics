package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/audit"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"go.uber.org/zap"
)

// UpdateMetricHandlerJSON — обработчик POST /update (JSON body).
// Принимает метрику в JSON-формате, обновляет её и возвращает обновлённую метрику.
func (mh *MetricHandler) UpdateMetricHandlerJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Info("got request with bad method", zap.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed. Expected POST")
		return
	}

	// Десериализуем тело запроса в структуру модели
	logger.Log.Debug("decoding request")
	var metric models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "Invalid request JSON body")
		return
	}

	// Проверяем корректность типа метрики
	if metric.MType != models.Counter && metric.MType != models.Gauge {
		logger.Log.Info("unsupported request type", zap.String("type", string(metric.MType)))
		writeJSONError(w, http.StatusUnprocessableEntity, "Unsupported request type")
		return
	}

	// Обновляем метрику через сервис
	if err := mh.service.UpdateMetric(r.Context(), metric); err != nil {
		logger.Log.Info("cannot update metric", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "Cannot update metric")
		return
	}
	if mh.publisher != nil {
		mh.publisher.Notify(audit.AuditEvent{
			Timestamp: time.Now().Unix(),
			Metrics:   []string{metric.ID},
			IPAddress: extractIPAddress(r),
		})
	}
	writeJSON(w, http.StatusOK, metric)
}

// GetMetricValueJSON — обработчик POST /value (JSON body).
// Принимает запрос с ID и типом метрики, возвращает её текущее значение в JSON.
func (mh *MetricHandler) GetMetricValueJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Info("got request with bad method", zap.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed. Expected POST")
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if parts[0] != "value" {
		logger.Log.Info("invalid path format", zap.String("path", path))
		writeJSONError(w, http.StatusNotFound, "Invalid path format")
		return
	}

	var metric models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "Invalid request JSON body")
		return
	}
	logger.Log.Debug("got request metric", zap.String("metric", metric.String()))

	// Проверяем корректность типа метрики
	if metric.MType != models.Counter && metric.MType != models.Gauge {
		logger.Log.Info("unsupported metric type", zap.String("type", string(metric.MType)))
		writeJSONError(w, http.StatusUnprocessableEntity, "Unsupported metric type")
		return
	}

	// Проверяем наличие ID метрики
	if metric.ID == "" {
		logger.Log.Info("got request with empty metric ID")
		writeJSONError(w, http.StatusNotFound, "Metric name is required")
		return
	}

	switch metric.MType {
	case models.Gauge:
		// Получаем значение Gauge метрики из сервиса
		value, err := mh.service.GetGauge(r.Context(), metric.ID)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "Metric not found")
			return
		}
		// Устанавливаем полученное значение в структуру метрики и отправляем JSON ответ
		metric.Value = &value
		writeJSON(w, http.StatusOK, metric)
	case models.Counter:
		// Получаем значение Counter метрики из сервиса
		value, err := mh.service.GetCounter(r.Context(), metric.ID)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "Metric not found")
			return
		}
		// Устанавливаем полученное значение в структуру метрики и отправляем JSON ответ
		metric.Delta = &value
		writeJSON(w, http.StatusOK, metric)
	default:
		writeJSONError(w, http.StatusUnprocessableEntity, "Unsupported metric type")
	}
}

// UpdateMetricsHandlerJSON — обработчик POST /updates (JSON body).
// Принимает массив метрик в JSON-формате и выполняет пакетное обновление.
func (mh *MetricHandler) UpdateMetricsHandlerJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Info("got request with bad method", zap.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed. Expected POST")
		return
	}

	// Десериализуем тело запроса в структуру модели
	logger.Log.Debug("decoding request")
	var metrics []models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "Invalid request JSON body")
		return
	}

	// Обновляем метрику через сервис
	if err := mh.service.UpdateMetrics(r.Context(), metrics); err != nil {
		if errors.Is(err, service.ErrUnknownMetricType) {
			logger.Log.Info("unsupported request type", zap.String("err", err.Error()))
			writeJSONError(w, http.StatusUnprocessableEntity, "Unsupported request type")
			return
		}
		logger.Log.Info("cannot update metrics", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "Cannot update metrics")
		return
	}

	if mh.publisher != nil {
		// Формируем список имён метрик для аудита
		metricNames := make([]string, 0, len(metrics))
		for _, m := range metrics {
			metricNames = append(metricNames, m.ID)
		}

		mh.publisher.Notify(audit.AuditEvent{
			Timestamp: time.Now().Unix(),
			Metrics:   metricNames,
			IPAddress: extractIPAddress(r),
		})
	}
	writeJSON(w, http.StatusOK, metrics)
}
