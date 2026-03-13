package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"go.uber.org/zap"
)

// EnableSyncSave включает синхронную запись после каждого обновления метрики
func (ms *MemStorage) EnableSyncSave(path string) {
	ms.syncSavePath = path
}

func (ms *MemStorage) saveIfEnabled() error {
	if ms.syncSavePath == "" {
		return nil
	}
	return ms.SaveToFile(ms.syncSavePath)
}

func (ms *MemStorage) SaveToFile(path string) error {
	if path == "" {
		return errors.New("file path is empty")
	}
	ctx := context.Background()
	gauges, err := ms.GetAllGauges(ctx)
	if err != nil {
		return fmt.Errorf("get all gauges: %w", err)
	}
	counters, err := ms.GetAllCounters(ctx)
	if err != nil {
		return fmt.Errorf("get all counters: %w", err)
	}

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

	ms.mu.Lock()
	tmpSyncSavePath := ms.syncSavePath
	if tmpSyncSavePath != "" {
		ms.syncSavePath = ""
	}
	ms.mu.Unlock()

	if tmpSyncSavePath != "" {
		defer func() {
			ms.mu.Lock()
			ms.syncSavePath = tmpSyncSavePath
			ms.mu.Unlock()
		}()
	}

	ctx := context.Background()

	var errs []error
	for _, m := range metrics {
		err := ms.loadMetric(ctx, m)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return false, errors.Join(errs...)
	}

	return true, nil
}

func (ms *MemStorage) loadMetric(ctx context.Context, metric models.Metrics) error {
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric must have a value")
		}
		if err := ms.UpdateGauge(ctx, metric.ID, *metric.Value); err != nil {
			return err
		}
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric must have a value")
		}
		if err := ms.UpdateCounter(ctx, metric.ID, *metric.Delta); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}
	return nil
}
