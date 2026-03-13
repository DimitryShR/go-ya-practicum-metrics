package handler_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricService - мок сервиса метрик
type MockMetricService struct {
	mock.Mock
}

func (m *MockMetricService) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricService) GetGauge(ctx context.Context, name string) (float64, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricService) GetCounter(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricService) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockMetricService) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]int64), args.Error(1)
}

func TestMetricHandler_UpdateMetricHandler(t *testing.T) {

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
				m.On("UpdateMetric", mock.Anything, gaugeMetric).Return(errors.New("Some error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Some error",
		},
		{
			name:           "Metric not found in context",
			metric:         models.Metrics{},
			mockSetup:      func(m *MockMetricService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок сервиса
			mockService := new(MockMetricService)
			tt.mockSetup(mockService)

			// Создаем handler
			metricHandler := handler.NewMetricHandler(mockService)

			// Создаем тестовый HTTP запрос
			req := httptest.NewRequest(http.MethodPost, "/update/", nil)

			// Добавляем метрику в контекст, если это не тест на отсутствие метрики
			if tt.name != "Metric not found in context" {
				ctx := context.WithValue(req.Context(), middleware.Metric, tt.metric)
				req = req.WithContext(ctx)
			}

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Вызываем handler
			metricHandler.UpdateMetricHandler(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Проверяем тело ответа в случае ошибки
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}

			// Проверяем заголовки в случае успеха
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			}

			// Проверяем, что все ожидаемые вызовы мока были выполнены
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricHandler_GetMetricValue_Errors(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		mockSetup      func(*MockMetricService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Gauge not found",
			metricType: "gauge",
			mockSetup: func(m *MockMetricService) {
				m.On("GetGauge", mock.Anything, "testMetric").Return(0.0, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric not found",
		},
		{
			name:       "Counter not found",
			metricType: "counter",
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

			req := httptest.NewRequest(http.MethodGet, "/value/"+tt.metricType+"/testMetric", nil)
			rr := httptest.NewRecorder()

			metricHandler.GetMetricValue(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}
