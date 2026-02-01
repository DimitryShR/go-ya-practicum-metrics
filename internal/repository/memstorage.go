package repository

import (
	"fmt"

	models "github.com/DimitryShR/go-ya-practicum-metrics/internal/model"
)

type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (ms *MemStorage) UpdateCounter(name string, value *int64) {
	ms.counters[name] += *value
}

func (ms *MemStorage) UpdateGauge(name string, value *float64) {
	ms.gauges[name] = *value
}

func (ms *MemStorage) UpdateMetric(metric models.Metrics) error {

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric must have a value")
		}
		ms.UpdateGauge(metric.ID, metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric must have a value")
		}
		ms.UpdateCounter(metric.ID, metric.Delta)
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}
	return nil
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	value, ok := ms.counters[name]
	return value, ok
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	value, ok := ms.gauges[name]
	return value, ok
}

func (ms *MemStorage) GetAllGauges() map[string]float64 {
	copied := make(map[string]float64, len(ms.gauges))
	for k, v := range ms.gauges {
		copied[k] = v
	}
	return copied
}

func (ms *MemStorage) GetAllCounters() map[string]int64 {
	copied := make(map[string]int64, len(ms.counters))
	for k, v := range ms.counters {
		copied[k] = v
	}
	return copied
}
