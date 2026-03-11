package service

type storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}
