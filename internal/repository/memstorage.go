package repository

import (
	"context"
	"fmt"
	"sync"
)

type MemStorage struct {
	mu           sync.RWMutex
	counters     map[string]int64
	gauges       map[string]float64
	syncSavePath string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (ms *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	ms.mu.Lock()
	ms.counters[name] += value
	ms.mu.Unlock()
	return ms.saveIfEnabled(ctx)
}

func (ms *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	ms.mu.Lock()
	ms.gauges[name] = value
	ms.mu.Unlock()
	return ms.saveIfEnabled(ctx)
}

func (ms *MemStorage) updateCounters(ctx context.Context, metrics map[string]int64) error {
	for metric, value := range metrics {
		if err := ms.UpdateCounter(ctx, metric, value); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemStorage) updateGauges(ctx context.Context, metrics map[string]float64) error {
	for metric, value := range metrics {
		if err := ms.UpdateGauge(ctx, metric, value); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemStorage) UpdateMetrics(ctx context.Context, counters map[string]int64, gauges map[string]float64) error {
	if len(counters) > 0 {
		if err := ms.updateCounters(ctx, counters); err != nil {
			return err
		}
	}
	if len(gauges) > 0 {
		if err := ms.updateGauges(ctx, gauges); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	if err := requireContext(ctx); err != nil {
		return 0, err
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.counters[name]
	if !ok {
		return 0, fmt.Errorf("counter metric not found: %s", name)
	}
	return value, nil
}

func (ms *MemStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	if err := requireContext(ctx); err != nil {
		return 0, err
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.gauges[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric not found: %s", name)
	}
	return value, nil
}

func (ms *MemStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	copied := make(map[string]float64, len(ms.gauges))
	for k, v := range ms.gauges {
		copied[k] = v
	}
	return copied, nil
}

func (ms *MemStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	copied := make(map[string]int64, len(ms.counters))
	for k, v := range ms.counters {
		copied[k] = v
	}
	return copied, nil
}
