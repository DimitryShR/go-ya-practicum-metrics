package agent

import (
	"testing"
)

func BenchmarkMetricsCollector_GetMetricsForReport(b *testing.B) {
	collector := NewMetricsCollector()
	// Заполняем метрики
	collector.CollectRuntime()
	collector.CollectRuntime()

	b.ReportAllocs()
	for b.Loop() {
		_ = collector.GetMetricsForReport()
	}
}

func BenchmarkMetricsCollector_CollectAndReport(b *testing.B) {
	collector := NewMetricsCollector()
	b.ReportAllocs()
	for b.Loop() {
		collector.CollectRuntime()
		_ = collector.GetMetricsForReport()
	}
}
