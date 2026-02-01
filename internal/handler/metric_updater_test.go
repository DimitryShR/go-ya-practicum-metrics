package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	middleware "github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricService - мок сервиса метрик
type MockMetricService struct {
	mock.Mock
}

func (m *MockMetricService) UpdateMetric(metric models.Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockMetricService) GetGauge(name string) (float64, bool) {
	args := m.Called(name)
	return args.Get(0).(float64), args.Bool(0)
}

func (m *MockMetricService) GetCounter(name string) (int64, bool) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Bool(0)
}

func (m *MockMetricService) GetAllGauges() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func (m *MockMetricService) GetAllCounters() map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]int64)
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
				m.On("UpdateMetric", gaugeMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success update counter metric",
			metric: counterMetric,
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetric", counterMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Service returns error",
			metric: gaugeMetric,
			mockSetup: func(m *MockMetricService) {
				m.On("UpdateMetric", gaugeMetric).Return(errors.New("Some error"))
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
