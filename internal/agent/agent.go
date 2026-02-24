package agent

import (
	"context"
	"log"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
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
	pollTicker := time.NewTicker(a.config.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(a.config.ReportInterval)
	defer reportTicker.Stop()

	a.logStartup()

	for {
		select {
		case <-ctx.Done():
			log.Println("Agent stopped")
			return
		case <-pollTicker.C:
			a.collectMetrics()
		case <-reportTicker.C:
			a.reportAllMetricJson()
		}
	}
}

func (a *Agent) logStartup() {
	log.Println("Agent started")
	log.Printf("Poll interval: %v", a.config.PollInterval)
	log.Printf("Report interval: %v", a.config.ReportInterval)
	log.Printf("Server address: %s", a.config.ServerAddress)
}

func (a *Agent) collectMetrics() {
	a.collector.Collect()
}

func (a *Agent) reportAllMetricJson() {
	metrics := a.collector.GetMetricsForReport()

	if err := a.client.SendAllMetricJson(metrics); err != nil {
		log.Printf("Failed to send metrics: %v", err)
		return
	}

	log.Printf("Successfully sent %d metrics", len(metrics))
}

func (a *Agent) reportMetrics() {
	metrics := a.collector.GetMetricsForReport()

	if err := a.client.SendMetrics(metrics); err != nil {
		log.Printf("Failed to send metrics: %v", err)
		return
	}

	log.Printf("Successfully sent %d metrics", len(metrics))
}
