// Package logger предоставляет глобальный логгер на основе zap.Logger.
package logger

import "go.uber.org/zap"

// Log — глобальный логгер. По умолчанию используется no-op логгер.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует глобальный логгер с указанным уровнем логирования.
// Уровень задаётся строкой: "info", "debug", "error".
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
