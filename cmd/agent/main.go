package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/agent"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	// Создаем контекст с обработкой сигналов
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)
	defer stop()

	// Загружаем конфигурацию
	cfg := config.NewAgentConfig()

	// Создаем и запускаем агент
	ag := agent.NewAgent(cfg)
	ag.Run(ctx)
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Fprintln(os.Stdout, "Build version:", buildVersion)
	fmt.Fprintln(os.Stdout, "Build date:", buildDate)
	fmt.Fprintln(os.Stdout, "Build commit:", buildCommit)
}
