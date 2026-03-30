package agent

import (
	"errors"
	"runtime"
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var runtimeGaugeMetrics = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
	"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
	"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
	"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
	"Sys", "TotalAlloc", "RandomValue",
}

// Проверяем сбор метрик и их корректность
func TestMetricsCollector_Collect(t *testing.T) {
	tests := []struct {
		name         string
		collectCount int
		wantMetrics  []string
	}{
		{
			name:         "single collect call",
			collectCount: 1,
			wantMetrics:  runtimeGaugeMetrics,
		},
		{
			name:         "multiple collect calls",
			collectCount: 3,
			wantMetrics:  runtimeGaugeMetrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector()

			// Вызываем Collect указанное количество раз
			for i := 0; i < tt.collectCount; i++ {
				collector.Collect()
			}

			// Получаем метрики для отчета
			metrics := collector.GetMetricsForReport()

			// Проверяем, что количество метрик соответствует ожидаемому
			// +1 для PollCount
			require.Len(t, metrics, len(tt.wantMetrics)+1, "unexpected number of metrics")

			// Создаем map для проверки наличия метрик
			metricsMap := make(map[string]bool)
			for _, m := range metrics {
				metricsMap[m.ID] = true
			}

			// Проверяем наличие всех ожидаемых gauge метрик
			for _, metricName := range tt.wantMetrics {
				assert.True(t, metricsMap[metricName], "metric %s not found", metricName)
			}

			// Проверяем наличие PollCount
			assert.True(t, metricsMap["PollCount"], "PollCount metric not found")

			// Проверяем тип метрик
			for _, m := range metrics {
				if m.ID == "PollCount" {
					assert.Equal(t, models.Counter, m.MType, "PollCount should be counter type")
					assert.NotNil(t, m.Delta, "PollCount should have Delta value")
					assert.Nil(t, m.Value, "PollCount should not have Value")

					// Проверяем значение PollCount
					if m.Delta != nil {
						assert.Equal(t, int64(tt.collectCount), *m.Delta,
							"PollCount should equal number of collect calls")
					}
				} else {
					assert.Equal(t, models.Gauge, m.MType, "metric %s should be gauge type", m.ID)
					assert.NotNil(t, m.Value, "gauge metric %s should have Value", m.ID)
					assert.Nil(t, m.Delta, "gauge metric %s should not have Delta", m.ID)
				}
			}

			// Проверяем, что после GetMetricsForReport счетчик сбрасывается
			metricsAfterReset := collector.GetMetricsForReport()
			for _, m := range metricsAfterReset {
				if m.ID == "PollCount" && m.Delta != nil {
					assert.Equal(t, int64(0), *m.Delta,
						"PollCount should be reset to 0 after GetMetricsForReport")
				}
			}
		})
	}
}

// Тестируем, что метрики runtime действительно собираются и изменяются
func TestMetricsCollector_Collect_RuntimeMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	// Первый сбор метрик
	collector.Collect()
	metrics1 := collector.GetMetricsForReport()

	// Второй сбор метрик
	collector.Collect()
	metrics2 := collector.GetMetricsForReport()

	// Создаем map для быстрого поиска
	metrics1Map := make(map[string]float64)
	for _, m := range metrics1 {
		if m.Value != nil {
			metrics1Map[m.ID] = *m.Value
		}
	}

	metrics2Map := make(map[string]float64)
	for _, m := range metrics2 {
		if m.Value != nil {
			metrics2Map[m.ID] = *m.Value
		}
	}

	// Проверяем, что значения метрик runtime изменились (хотя бы некоторые)
	changedCount := 0
	for name, value1 := range metrics1Map {
		if name == "RandomValue" {
			continue // RandomValue всегда разный
		}
		if value2, ok := metrics2Map[name]; ok {
			if value1 != value2 {
				changedCount++
			}
		}
	}

	// Ожидаем, что хотя бы некоторые метрики runtime изменятся между вызовами
	assert.Greater(t, changedCount, 0,
		"expected at least some runtime metrics to change between collects")
}

// Проверяем корректность инициализации MetricsCollector
func TestMetricsCollector_NewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.metrics)
	assert.NotNil(t, collector.readMemStats)
	assert.NotNil(t, collector.readVirtualMemory)
	assert.NotNil(t, collector.readCPUPercent)
	assert.Equal(t, 0, len(collector.metrics))
	assert.Equal(t, int64(0), collector.pollCount)
}

func TestMetricsCollector_CollectSystem(t *testing.T) {
	collector := NewMetricsCollector()
	collector.readVirtualMemory = func() (*virtualMemoryStat, error) {
		return &virtualMemoryStat{
			Total: 4096,
			Free:  1024,
		}, nil
	}
	collector.readCPUPercent = func() ([]float64, error) {
		return []float64{10.5, 20.25, 30.75}, nil
	}

	err := collector.CollectSystem()
	require.NoError(t, err)

	metrics := collector.GetMetricsForReport()
	metricsMap := make(map[string]models.Metrics, len(metrics))
	for _, metric := range metrics {
		metricsMap[metric.ID] = metric
	}

	require.Contains(t, metricsMap, "TotalMemory")
	require.Contains(t, metricsMap, "FreeMemory")
	require.Contains(t, metricsMap, "CPUutilization1")
	require.Contains(t, metricsMap, "CPUutilization2")
	require.Contains(t, metricsMap, "CPUutilization3")

	assert.Equal(t, 4096.0, *metricsMap["TotalMemory"].Value)
	assert.Equal(t, 1024.0, *metricsMap["FreeMemory"].Value)
	assert.Equal(t, 10.5, *metricsMap["CPUutilization1"].Value)
	assert.Equal(t, 20.25, *metricsMap["CPUutilization2"].Value)
	assert.Equal(t, 30.75, *metricsMap["CPUutilization3"].Value)
}

func TestMetricsCollector_CollectSystem_UpdatesCPUSet(t *testing.T) {
	collector := NewMetricsCollector()
	collector.readVirtualMemory = func() (*virtualMemoryStat, error) {
		return &virtualMemoryStat{Total: 1, Free: 1}, nil
	}
	collector.readCPUPercent = func() ([]float64, error) {
		return []float64{1, 2, 3}, nil
	}

	require.NoError(t, collector.CollectSystem())

	collector.readCPUPercent = func() ([]float64, error) {
		return []float64{4, 5}, nil
	}
	require.NoError(t, collector.CollectSystem())

	metrics := collector.GetMetricsForReport()
	metricsMap := make(map[string]models.Metrics, len(metrics))
	for _, metric := range metrics {
		metricsMap[metric.ID] = metric
	}

	require.Contains(t, metricsMap, "CPUutilization1")
	require.Contains(t, metricsMap, "CPUutilization2")
	assert.NotContains(t, metricsMap, "CPUutilization3")
	assert.Equal(t, 4.0, *metricsMap["CPUutilization1"].Value)
	assert.Equal(t, 5.0, *metricsMap["CPUutilization2"].Value)
}

func TestMetricsCollector_CollectSystem_ReturnsJoinedError(t *testing.T) {
	collector := NewMetricsCollector()
	memErr := errors.New("memory failed")
	cpuErr := errors.New("cpu failed")
	collector.readVirtualMemory = func() (*virtualMemoryStat, error) {
		return nil, memErr
	}
	collector.readCPUPercent = func() ([]float64, error) {
		return nil, cpuErr
	}

	err := collector.CollectSystem()
	require.Error(t, err)
	assert.ErrorIs(t, err, memErr)
	assert.ErrorIs(t, err, cpuErr)
}

func TestMetricsCollector_CollectSystem_DefaultCPUCount(t *testing.T) {
	collector := NewMetricsCollector()

	err := collector.CollectSystem()
	if err != nil {
		t.Skipf("gopsutil is unavailable in current environment: %v", err)
	}

	metrics := collector.GetMetricsForReport()
	cpuMetricCount := 0
	for _, metric := range metrics {
		if len(metric.ID) >= len("CPUutilization") && metric.ID[:len("CPUutilization")] == "CPUutilization" {
			cpuMetricCount++
		}
	}

	assert.Equal(t, runtime.NumCPU(), cpuMetricCount)
}
