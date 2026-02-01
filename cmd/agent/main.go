package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	agent "github.com/DimitryShR/go-ya-practicum-metrics/internal/agent"
	config "github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
)

func main() {
	// Создаем контекст с обработкой сигналов
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)
	defer stop()

	// Загружаем конфигурацию
	cfg := config.NewDefaultAgentConfig()

	// Создаем и запускаем агент
	ag := agent.NewAgent(cfg)
	ag.Run(ctx)
}
