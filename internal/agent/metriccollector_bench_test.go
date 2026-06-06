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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.GetMetricsForReport()
	}
}

func BenchmarkMetricsCollector_CollectAndReport(b *testing.B) {
	collector := NewMetricsCollector()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.CollectRuntime()
		_ = collector.GetMetricsForReport()
	}
}
