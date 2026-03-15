package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMetricHandler_UpdateMetricHandlerJSON(t *testing.T) {

	// Создаем тестовые метрики
	gaugeMetric := models.Metrics{
		ID:    "testGauge",
		MType: "gauge",
		Value: func() *float64 { v := 10.5; return &v }(),
	}

	counterMetric := models.Metrics{
		ID:    "testCounter",
		MType: "counter",
		Delta: func() *int64 { v := int64(100); return &v }(),
	}

	tests := []struct {
		name           string
		metric         models.Metrics
		rawBody        string
		mockSetup      func(*MockMetricService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "Success update gauge metric",
			metric: gaugeMetric,
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetric", mock.Anything, gaugeMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success update counter metric",
			metric: counterMetric,
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetric", mock.Anything, counterMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Service returns error",
			metric: gaugeMetric,
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetric", mock.Anything, gaugeMetric).Return(errors.New("Cannot update metric"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Cannot update metric",
		},
		{
			name:           "Metric not found",
			metric:         models.Metrics{},
			mockSetup:      func(m *MockMetricService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Unsupported request type",
		},
		{
			name:           "Invalid JSON body",
			rawBody:        `{"id":"testGauge","type":"gauge"`,
			mockSetup:      func(m *MockMetricService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request JSON body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок сервиса
			mockService := new(MockMetricService)
			tt.mockSetup(mockService)

			// Создаем handler
			metricHandler := handler.NewMetricHandler(mockService)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = &bytes.Buffer{}
				if err := json.NewEncoder(reqBody).Encode(tt.metric); err != nil {
					t.Fatalf("Failed to encode metric to JSON: %v", err)
				}
			}

			// Создаем тестовый HTTP запрос
			req := httptest.NewRequest(http.MethodPost, "/update/", reqBody)

			// Устанавливаем заголовок Content-Type для JSON
			req.Header.Set("Content-Type", "application/json")

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Вызываем handler
			metricHandler.UpdateMetricHandlerJSON(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Проверяем тело ответа в случае ошибки
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}

			// Проверяем заголовки в случае успеха
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}

			// Проверяем, что все ожидаемые вызовы мока были выполнены
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricHandler_GetMetricValueJSON_Errors(t *testing.T) {
	tests := []struct {
		name           string
		metric         models.Metrics
		mockSetup      func(*MockMetricService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Gauge not found",
			metric: models.Metrics{
				ID:    "testMetric",
				MType: "gauge",
			},
			mockSetup: func(m *MockMetricService) {
				m.On("GetGauge", mock.Anything, "testMetric").Return(0.0, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric not found",
		},
		{
			name: "Counter not found",
			metric: models.Metrics{
				ID:    "testMetric",
				MType: "counter",
			},
			mockSetup: func(m *MockMetricService) {
				m.On("GetCounter", mock.Anything, "testMetric").Return(int64(0), sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricService)
			tt.mockSetup(mockService)

			metricHandler := handler.NewMetricHandler(mockService)

			var reqBody bytes.Buffer
			if err := json.NewEncoder(&reqBody).Encode(tt.metric); err != nil {
				t.Fatalf("Failed to encode metric to JSON: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/value/", &reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			metricHandler.GetMetricValueJSON(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricHandler_UpdateMetricsHandlerJSON(t *testing.T) {
	gaugeMetric := models.Metrics{
		ID:    "testGauge",
		MType: "gauge",
		Value: func() *float64 { v := 10.5; return &v }(),
	}

	counterMetric := models.Metrics{
		ID:    "testCounter",
		MType: "counter",
		Delta: func() *int64 { v := int64(100); return &v }(),
	}

	tests := []struct {
		name           string
		method         string
		metrics        []models.Metrics
		rawBody        string
		mockSetup      func(*MockMetricService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Success update metrics",
			method:  http.MethodPost,
			metrics: []models.Metrics{gaugeMetric, counterMetric},
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetrics", mock.Anything, []models.Metrics{gaugeMetric, counterMetric}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			method:         http.MethodPost,
			rawBody:        `[{"id":"testGauge","type":"gauge"}`,
			mockSetup:      func(m *MockMetricService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request JSON body",
		},
		{
			name:    "Unsupported metric type",
			method:  http.MethodPost,
			metrics: []models.Metrics{{ID: "badMetric", MType: "unknown"}},
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetrics", mock.Anything, []models.Metrics{{ID: "badMetric", MType: "unknown"}}).
					Return(service.UnknownMetricTypeErr)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Unsupported request type",
		},
		{
			name:    "Service returns error",
			method:  http.MethodPost,
			metrics: []models.Metrics{gaugeMetric},
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetrics", mock.Anything, []models.Metrics{gaugeMetric}).
					Return(errors.New("Cannot update metrics"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Cannot update metrics",
		},
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			metrics:        []models.Metrics{gaugeMetric},
			mockSetup:      func(m *MockMetricService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricService)
			tt.mockSetup(mockService)

			metricHandler := handler.NewMetricHandler(mockService)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = &bytes.Buffer{}
				if err := json.NewEncoder(reqBody).Encode(tt.metrics); err != nil {
					t.Fatalf("Failed to encode metrics to JSON: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/updates", reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			metricHandler.UpdateMetricsHandlerJSON(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}

			mockService.AssertExpectations(t)
		})
	}
}
