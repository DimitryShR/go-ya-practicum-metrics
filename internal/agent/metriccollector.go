package agent

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// virtualMemoryStat — упрощённая структура для виртуальной памяти (для тестирования).
type virtualMemoryStat struct {
	Total uint64
	Free  uint64
}

// MetricsCollector собирает runtime и системные метрики.
type MetricsCollector struct {
	mu                sync.Mutex
	metrics           map[string]float64
	pollCount         int64
	cpuMetricCount    int
	readMemStats      func(*runtime.MemStats)
	readVirtualMemory func() (*virtualMemoryStat, error)
	readCPUPercent    func() ([]float64, error)
}

// NewMetricsCollector создаёт новый сборщик метрик с функциями для чтения runtime/системных данных.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:      make(map[string]float64),
		readMemStats: runtime.ReadMemStats,
		readVirtualMemory: func() (*virtualMemoryStat, error) {
			vm, err := mem.VirtualMemory()
			if err != nil {
				return nil, err
			}
			return &virtualMemoryStat{
				Total: vm.Total,
				Free:  vm.Free,
			}, nil
		},
		readCPUPercent: func() ([]float64, error) {
			return cpu.Percent(0, true)
		},
	}
}

// Collect собирает runtime метрики (алиас для CollectRuntime).
func (c *MetricsCollector) Collect() {
	c.CollectRuntime()
}

// CollectRuntime собирает runtime метрики (MemStats) и обновляет внутреннее хранилище.
func (c *MetricsCollector) CollectRuntime() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Увеличиваем счетчик опросов
	c.pollCount++

	// Собираем runtime метрики
	var memStats runtime.MemStats

	c.readMemStats(&memStats)

	// Gauge метрики из runtime
	c.metrics["Alloc"] = float64(memStats.Alloc)
	c.metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	c.metrics["Frees"] = float64(memStats.Frees)
	c.metrics["GCCPUFraction"] = memStats.GCCPUFraction
	c.metrics["GCSys"] = float64(memStats.GCSys)
	c.metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	c.metrics["HeapIdle"] = float64(memStats.HeapIdle)
	c.metrics["HeapInuse"] = float64(memStats.HeapInuse)
	c.metrics["HeapObjects"] = float64(memStats.HeapObjects)
	c.metrics["HeapReleased"] = float64(memStats.HeapReleased)
	c.metrics["HeapSys"] = float64(memStats.HeapSys)
	c.metrics["LastGC"] = float64(memStats.LastGC)
	c.metrics["Lookups"] = float64(memStats.Lookups)
	c.metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	c.metrics["MCacheSys"] = float64(memStats.MCacheSys)
	c.metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	c.metrics["MSpanSys"] = float64(memStats.MSpanSys)
	c.metrics["Mallocs"] = float64(memStats.Mallocs)
	c.metrics["NextGC"] = float64(memStats.NextGC)
	c.metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	c.metrics["NumGC"] = float64(memStats.NumGC)
	c.metrics["OtherSys"] = float64(memStats.OtherSys)
	c.metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	c.metrics["StackInuse"] = float64(memStats.StackInuse)
	c.metrics["StackSys"] = float64(memStats.StackSys)
	c.metrics["Sys"] = float64(memStats.Sys)
	c.metrics["TotalAlloc"] = float64(memStats.TotalAlloc)

	// Дополнительная gauge метрика
	c.metrics["RandomValue"] = rand.Float64()
}

// CollectSystem собирает системные метрики (CPU, RAM через gopsutil) и обновляет хранилище.
func (c *MetricsCollector) CollectSystem() error {
	virtualMemory, memoryErr := c.readVirtualMemory()
	cpuUtilization, cpuErr := c.readCPUPercent()

	c.mu.Lock()
	defer c.mu.Unlock()

	if memoryErr == nil {
		c.metrics["TotalMemory"] = float64(virtualMemory.Total)
		c.metrics["FreeMemory"] = float64(virtualMemory.Free)
	}

	if cpuErr == nil {
		// Удаляем метрики для ядер, которых больше нет.
		// c.cpuMetricCount – количество ядер, зафиксированное в прошлый раз.
		// len(cpuUtilization) – актуальное количество ядер.
		for idx := c.cpuMetricCount; idx > len(cpuUtilization); idx-- {
			delete(c.metrics, fmt.Sprintf("CPUutilization%d", idx))
		}
		// Добавляем или обновляем метрики для текущих ядер.
		for idx, utilization := range cpuUtilization {
			c.metrics[fmt.Sprintf("CPUutilization%d", idx+1)] = utilization
		}
		// Фиксируем актуальное количество ядер
		c.cpuMetricCount = len(cpuUtilization)
	}

	var errs []error
	if memoryErr != nil {
		errs = append(errs, fmt.Errorf("failed to collect memory metrics: %w", memoryErr))
	}
	if cpuErr != nil {
		errs = append(errs, fmt.Errorf("failed to collect CPU metrics: %w", cpuErr))
	}
	return errors.Join(errs...)
}

// GetMetricsForReport возвращает все накопленные метрики для отправки на сервер.
// Сбрасывает внутренний счётчик PollCount после сбора.
func (c *MetricsCollector) GetMetricsForReport() []models.Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Предварительное выделение слайса необходимой ёмкости
	metrics := make([]models.Metrics, 0, len(c.metrics)+1)

	// Добавляем все gauge метрики
	for name, value := range c.metrics {
		val := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &val,
		})
	}

	// Добавляем counter метрику PollCount
	delta := c.pollCount
	metrics = append(metrics, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &delta,
	})

	// Сбрасываем счетчик после сбора метрик для отчета
	c.pollCount = 0
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics
}
