package agent

import (
	"context"
	"log"
	"time"

	config "github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
)

type Agent struct {
	config    *config.AgentConfig
	collector *MetricsCollector
	client    *MetricsClient
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{
		config:    cfg,
		collector: NewMetricsCollector(),
		client:    NewMetricsClient(cfg),
	}
}

func (a *Agent) Run(ctx context.Context) {
	// Каналы для тикеров
	pollTicker := time.NewTicker(a.config.PollInterval)
	reportTicker := time.NewTicker(a.config.ReportInterval)

	log.Println("Agent started")
	log.Printf("Poll interval: %v", a.config.PollInterval)
	log.Printf("Report interval: %v", a.config.ReportInterval)
	log.Printf("Server address: %s", a.config.ServerAddress)

	for {
		select {
		case <-ctx.Done():
			log.Println("Agent stopped")
			return

		case <-pollTicker.C:
			// Сбор метрик
			a.collector.Collect()

		case <-reportTicker.C:
			// Отправка метрик
			metrics := a.collector.GetMetricsForReport()

			if err := a.client.SendMetrics(metrics); err != nil {
				log.Printf("Failed to send metrics: %v", err)
			} else {
				log.Printf("Successfully sent %d metrics", len(metrics))
			}
		}
	}
}
