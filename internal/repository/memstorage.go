package repository

import (
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

func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.mu.Lock()
	ms.counters[name] += value
	ms.mu.Unlock()
	return ms.saveIfEnabled()
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.mu.Lock()
	ms.gauges[name] = value
	ms.mu.Unlock()
	return ms.saveIfEnabled()
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.counters[name]
	return value, ok
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.gauges[name]
	return value, ok
}

func (ms *MemStorage) GetAllGauges() map[string]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	copied := make(map[string]float64, len(ms.gauges))
	for k, v := range ms.gauges {
		copied[k] = v
	}
	return copied
}

func (ms *MemStorage) GetAllCounters() map[string]int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	copied := make(map[string]int64, len(ms.counters))
	for k, v := range ms.counters {
		copied[k] = v
	}
	return copied
}
