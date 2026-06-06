package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

// MetricService определяет интерфейс бизнес-логики работы с метриками.
type MetricService interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}

type metricService struct {
	repo Storage
}

// ErrUnknownMetricType возвращается при попытке обработать метрику с неизвестным типом.
var ErrUnknownMetricType = errors.New("unsupported metric type")

// NewMetricService создаёт новый сервис метрик с указанным хранилищем.
func NewMetricService(repo Storage) MetricService {
	return &metricService{repo: repo}
}

// UpdateMetric обновляет одну метрику в зависимости от её типа (gauge или counter).
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

// UpdateMetrics обновляет набор метрик, группируя их по типам для пакетной обработки.
func (s *metricService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {

	counters := make(map[string]int64, 0)
	gauges := make(map[string]float64, 0)

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge metric must have value")
			}
			gauges[string(metric.ID)] = *metric.Value
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("counter metric must have delta")
			}
			if _, ok := counters[string(metric.ID)]; !ok {
				counters[string(metric.ID)] = *metric.Delta
			} else {
				counters[string(metric.ID)] += *metric.Delta
			}
		default:
			return fmt.Errorf("%w: %v", ErrUnknownMetricType, metric.MType)
		}
	}

	if len(counters) > 0 || len(gauges) > 0 {
		if err := s.repo.UpdateMetrics(ctx, counters, gauges); err != nil {
			return err
		}
	}

	return nil
}

// GetGauge возвращает значение gauge-метрики по имени.
func (s *metricService) GetGauge(ctx context.Context, name string) (float64, error) {
	return s.repo.GetGauge(ctx, name)
}

// GetCounter возвращает значение counter-метрики по имени.
func (s *metricService) GetCounter(ctx context.Context, name string) (int64, error) {
	return s.repo.GetCounter(ctx, name)
}

// GetAllGauges возвращает все gauge-метрики в виде карты имя -> значение.
func (s *metricService) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	return s.repo.GetAllGauges(ctx)
}

// GetAllCounters возвращает все counter-метрики в виде карты имя -> значение.
func (s *metricService) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	return s.repo.GetAllCounters(ctx)
}
