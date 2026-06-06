package compress_test

import (
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/compress"
)

// Маленький JSON payload — одна метрика
var smallPayload = []byte(`{"id":"Alloc","type":"gauge","value":1234567.89}`)

// Большой JSON payload — пакет из 30 метрик
var largePayload = []byte(`[
  {"id":"Alloc","type":"gauge","value":1234567.89},
  {"id":"BuckHashSys","type":"gauge","value":2345678.90},
  {"id":"Frees","type":"gauge","value":3456789.01},
  {"id":"GCCPUFraction","type":"gauge","value":0.00123},
  {"id":"GCSys","type":"gauge","value":4567890.12},
  {"id":"HeapAlloc","type":"gauge","value":5678901.23},
  {"id":"HeapIdle","type":"gauge","value":6789012.34},
  {"id":"HeapInuse","type":"gauge","value":7890123.45},
  {"id":"HeapObjects","type":"gauge","value":8901234.56},
  {"id":"HeapReleased","type":"gauge","value":9012345.67},
  {"id":"HeapSys","type":"gauge","value":11223344.55},
  {"id":"LastGC","type":"gauge","value":22334455.66},
  {"id":"Lookups","type":"gauge","value":33445566.77},
  {"id":"MCacheInuse","type":"gauge","value":44556677.88},
  {"id":"MCacheSys","type":"gauge","value":55667788.99},
  {"id":"MSpanInuse","type":"gauge","value":66778899.00},
  {"id":"MSpanSys","type":"gauge","value":77889900.11},
  {"id":"Mallocs","type":"gauge","value":88990011.22},
  {"id":"NextGC","type":"gauge","value":99001122.33},
  {"id":"NumForcedGC","type":"gauge","value":11},
  {"id":"NumGC","type":"gauge","value":22},
  {"id":"OtherSys","type":"gauge","value":334455.66},
  {"id":"PauseTotalNs","type":"gauge","value":445566.77},
  {"id":"StackInuse","type":"gauge","value":556677.88},
  {"id":"StackSys","type":"gauge","value":667788.99},
  {"id":"Sys","type":"gauge","value":778899.00},
  {"id":"TotalAlloc","type":"gauge","value":889900.11},
  {"id":"PollCount","type":"counter","delta":1},
  {"id":"RandomValue","type":"gauge","value":0.789012345},
  {"id":"TotalMemory","type":"gauge","value":8589934592}
]`)

func BenchmarkGzipData_SmallPayload(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := compress.GzipData(smallPayload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGzipData_LargePayload(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := compress.GzipData(largePayload); err != nil {
			b.Fatal(err)
		}
	}
}
