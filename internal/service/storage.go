// Package service предоставляет бизнес-логику работы с метриками.
// Содержит интерфейсы Storage и MetricService, а также их реализации.
package service

import "context"

// Storage определяет интерфейс хранилища метрик.
// Реализации могут быть in-memory, файловыми или PostgreSQL.
type Storage interface {
	// UpdateGauge обновляет gauge-метрику с указанным именем и значением.
	UpdateGauge(ctx context.Context, name string, value float64) error
	// UpdateCounter обновляет counter-метрику с указанным именем и дельтой.
	UpdateCounter(ctx context.Context, name string, delta int64) error
	// UpdateMetrics выполняет пакетное обновление counter и gauge метрик.
	UpdateMetrics(ctx context.Context, counters map[string]int64, gauges map[string]float64) error
	// GetGauge возвращает значение gauge-метрики по имени.
	GetGauge(ctx context.Context, name string) (float64, error)
	// GetCounter возвращает значение counter-метрики по имени.
	GetCounter(ctx context.Context, name string) (int64, error)
	// GetAllGauges возвращает все gauge-метрики в виде карты имя -> значение.
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	// GetAllCounters возвращает все counter-метрики в виде карты имя -> значение.
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}
