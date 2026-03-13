package service

import "context"

type Storage interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, delta int64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}
