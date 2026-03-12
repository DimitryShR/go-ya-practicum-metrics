package service

import (
	"fmt"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

// MetricService определяет бизнес-логику работы с метриками
type MetricService interface {
	UpdateMetric(metric models.Metrics) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

type metricService struct {
	repo storage
}

func NewMetricService(repo storage) MetricService {
	return &metricService{repo: repo}
}

func (s *metricService) UpdateMetric(metric models.Metrics) error {
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric must have value")
		}
		return s.repo.UpdateGauge(metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric must have delta")
		}
		return s.repo.UpdateCounter(metric.ID, *metric.Delta)
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}
}

func (s *metricService) GetGauge(name string) (float64, bool) {
	// Возвращаем gauge по имени
	return s.repo.GetGauge(name)
}

func (s *metricService) GetCounter(name string) (int64, bool) {
	// Возвращаем counter по имени
	return s.repo.GetCounter(name)
}

func (s *metricService) GetAllGauges() map[string]float64 {
	// Возвращаем все gauge метрики
	return s.repo.GetAllGauges()
}

func (s *metricService) GetAllCounters() map[string]int64 {
	// Возвращаем все counter метрики
	return s.repo.GetAllCounters()
}
