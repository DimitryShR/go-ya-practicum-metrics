package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateMetric(metric models.Metrics) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

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

// EnableSyncSave включает синхронную запись после каждого обновления метрики
func (ms *MemStorage) EnableSyncSave(path string) {
	ms.syncSavePath = path
}

func (ms *MemStorage) UpdateCounter(name string, value *int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += *value
}

func (ms *MemStorage) UpdateGauge(name string, value *float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
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

	if ms.syncSavePath != "" {
		if err := ms.SaveToFile(ms.syncSavePath); err != nil {
			return fmt.Errorf("sync save failed: %w", err)
		}
	}
	return nil
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

func (ms *MemStorage) SaveToFile(path string) error {
	if path == "" {
		return errors.New("file path is empty")
	}

	gauges := ms.GetAllGauges()
	counters := ms.GetAllCounters()

	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))
	for name, value := range gauges {
		val := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &val,
		})
	}
	for name, delta := range counters {
		d := delta
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &d,
		})
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal storage snapshot: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "metrics-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp storage file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write storage snapshot: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp storage file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace storage file: %w", err)
	}

	return nil
}

func (ms *MemStorage) LoadFromFile(path string) (bool, error) {
	if path == "" {
		return false, errors.New("file path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Log.Warn("Storage file does not exist, starting with empty metrics", zap.String("file", path))
			return false, nil
		}
		return false, fmt.Errorf("read storage file: %w", err)
	}

	if len(data) == 0 {
		return false, nil
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return false, fmt.Errorf("unmarshal storage snapshot: %w", err)
	}

	for _, m := range metrics {
		ms.UpdateMetric(m)
	}

	return true, nil
}
