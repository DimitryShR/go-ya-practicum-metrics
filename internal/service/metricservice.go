package service

import (
	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
	repository "github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
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
	repo repository.Storage
}

func NewMetricService(repo repository.Storage) MetricService {
	return &metricService{repo: repo}
}

func (s *metricService) UpdateMetric(metric models.Metrics) error {
	// Здесь можно добавить бизнес-логику:
	// - валидацию данных
	// - логирование
	// - кэширование
	// - вычисление хешей и т.д.

	// Пока просто делегируем в репозиторий
	return s.repo.UpdateMetric(metric)
}

func (s *metricService) GetGauge(name string) (float64, bool) {
	// Бизнес-логика для получения gauge
	return s.repo.GetGauge(name)
}

func (s *metricService) GetCounter(name string) (int64, bool) {
	// Бизнес-логика для получения counter
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
