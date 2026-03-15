package agent

import (
	"context"
	"errors"
	"log"
	"net/http"
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
			a.reportBatchMetricsJSON()
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

func (a *Agent) reportBatchMetricsJSON() {
	metrics := a.collector.GetMetricsForReport()
	if len(metrics) == 0 {
		log.Printf("No metrics to send at this time")
		return
	}

	if err := a.client.BatchSendMetricsJSON(metrics); err != nil {
		log.Printf("Failed to send metrics to /updates: %v", err)

		// Для обратной совместимости, если метод /updates отсутствует
		var statusErr *HTTPStatusError
		if !errors.As(err, &statusErr) {
			return
		}
		switch statusErr.Code {
		case http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotImplemented:
			if err := a.client.SendAllMetricJSON(metrics); err != nil {
				log.Printf("Failed to send metrics to /update: %v", err)
				return
			}
		default:
			return
		}

	}
	log.Printf("Successfully sent %d metrics", len(metrics))

}

func (a *Agent) reportAllMetricJSON() {
	metrics := a.collector.GetMetricsForReport()

	if err := a.client.SendAllMetricJSON(metrics); err != nil {
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
