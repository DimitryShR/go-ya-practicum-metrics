package logger

import "go.uber.org/zap"

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	// парсим уровень логирования из строки, например "info", "debug", "error"
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём конфигурацию логгера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень логирования в конфигурацию
	cfg.Level = lvl
	// создаём логгер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// присваиваем глобальной переменной Log новый логгер
	Log = zl
	return nil
}
