package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"go.uber.org/zap"
)

// UpdateMetricHandlerJSON - обработчик POST /update (JSON тело)
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
		writeJSONError(w, http.StatusInternalServerError, "Cannot decode request JSON body")
		return
	}

	// Проверяем корректность типа метрики
	if metric.MType != models.Counter && metric.MType != models.Gauge {
		logger.Log.Info("unsupported request type", zap.String("type", string(metric.MType)))
		writeJSONError(w, http.StatusUnprocessableEntity, "Unsupported request type")
		return
	}

	// Обновляем метрику через сервис
	if err := mh.service.UpdateMetric(metric); err != nil {
		logger.Log.Info("cannot update metric", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "Cannot update metric")
		return
	}
	writeJSON(w, http.StatusOK, metric)
}

// GetMetricValueJSON - обработчик POST /value/ (JSON Body)
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

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		writeJSONError(w, http.StatusInternalServerError, "Cannot decode request JSON body")
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
		value, ok := mh.service.GetGauge(metric.ID)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "Metric not found")
			return
		}
		// Устанавливаем полученное значение в структуру метрики и отправляем JSON ответ
		metric.Value = &value
		writeJSON(w, http.StatusOK, metric)
	case models.Counter:
		// Получаем значение Counter метрики из сервиса
		value, ok := mh.service.GetCounter(metric.ID)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "Metric not found")
			return
		}
		// Устанавливаем полученное значение в структуру метрики и отправляем JSON ответ
		metric.Delta = &value
		writeJSON(w, http.StatusOK, metric)
	default:
		writeJSONError(w, http.StatusBadRequest, "Unsupported metric type")
	}
}
