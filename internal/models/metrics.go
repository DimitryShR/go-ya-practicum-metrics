package models

import "fmt"

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
	Hash  string     `json:"hash,omitempty"`
}

func (m Metrics) String() string {
	return fmt.Sprintf("Metrics{ID: %s, MType: %s, Delta: %v, Value: %v, Hash: %s}",
		m.ID, m.MType, m.Delta, m.Value, m.Hash)
}
