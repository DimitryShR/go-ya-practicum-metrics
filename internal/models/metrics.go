// Package models определяет типы данных для метрик, используемых в системе сбора метрик.
// Поддерживаются два типа метрик: gauge (вещественное число) и counter (счётчик).
package models

import "fmt"

// MetricType представляет тип метрики.
type MetricType string

const (
	// Counter — тип метрики-счётчика (int64). Используется для подсчёта событий.
	Counter MetricType = "counter"
	// Gauge — тип метрики-значения (float64). Используется для хранения произвольных числовых значений.
	Gauge MetricType = "gauge"
)

// Metrics представляет собой плоскую модель метрики.
//
// Delta и Value объявлены через указатели, чтобы отличать значение "0" от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	// ID — уникальное имя метрики.
	ID string `json:"id"`
	// MType — тип метрики ("counter" или "gauge").
	MType MetricType `json:"type"`
	// Delta — значение для counter-метрики (*int64, omitempty).
	Delta *int64 `json:"delta,omitempty"`
	// Value — значение для gauge-метрики (*float64, omitempty).
	Value *float64 `json:"value,omitempty"`
	// Hash — HMAC-SHA256 подпись метрики (опционально).
	Hash string `json:"hash,omitempty"`
}

// String возвращает строковое представление метрики.
func (m Metrics) String() string {
	return fmt.Sprintf("Metrics{ID: %s, MType: %s, Delta: %v, Value: %v, Hash: %s}",
		m.ID, m.MType, m.Delta, m.Value, m.Hash)
}
