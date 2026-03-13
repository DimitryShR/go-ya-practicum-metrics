package service

import (
	"context"
	"fmt"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

// MetricService определяет бизнес-логику работы с метриками
type MetricService interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}

type metricService struct {
	repo Storage
}

func NewMetricService(repo Storage) MetricService {
	return &metricService{repo: repo}
}

func (s *metricService) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric must have value")
		}
		return s.repo.UpdateGauge(ctx, metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric must have delta")
		}
		return s.repo.UpdateCounter(ctx, metric.ID, *metric.Delta)
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}
}

func (s *metricService) GetGauge(ctx context.Context, name string) (float64, error) {
	// Возвращаем gauge по имени
	return s.repo.GetGauge(ctx, name)
}

func (s *metricService) GetCounter(ctx context.Context, name string) (int64, error) {
	// Возвращаем counter по имени
	return s.repo.GetCounter(ctx, name)
}

func (s *metricService) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	// Возвращаем все gauge метрики
	return s.repo.GetAllGauges(ctx)
}

func (s *metricService) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	// Возвращаем все counter метрики
	return s.repo.GetAllCounters(ctx)
}
