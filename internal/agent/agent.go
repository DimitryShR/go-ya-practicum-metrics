package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/models"
)

const metricsQueueSize = 256

type metricsCollector interface {
	CollectRuntime()
	CollectSystem() error
	GetMetricsForReport() []models.Metrics
}

type metricsSender interface {
	SendMetricJSONWithContext(context.Context, models.Metrics) error
}

type Agent struct {
	config    *config.AgentConfig
	collector metricsCollector
	sender    metricsSender
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{
		config:    cfg,
		collector: NewMetricsCollector(),
		sender:    NewMetricsClient(cfg),
	}
}

func (a *Agent) Run(ctx context.Context) {
	a.logStartup()

	jobs := make(chan models.Metrics, metricsQueueSize)
	var wg sync.WaitGroup

	wg.Add(3 + a.config.RateLimit)
	go func() {
		defer wg.Done()
		a.runRuntimeCollector(ctx)
	}()
	go func() {
		defer wg.Done()
		a.runSystemCollector(ctx)
	}()
	go func() {
		defer wg.Done()
		a.runReporter(ctx, jobs)
		close(jobs)
	}()

	for workerID := 1; workerID <= a.config.RateLimit; workerID++ {
		go func(id int) {
			defer wg.Done()
			a.runSenderWorker(ctx, id, jobs)
		}(workerID)
	}

	wg.Wait()
	log.Println("Agent stopped")
}

func (a *Agent) logStartup() {
	log.Println("Agent started")
	log.Printf("Poll interval: %v", a.config.PollInterval)
	log.Printf("Report interval: %v", a.config.ReportInterval)
	log.Printf("Rate limit: %d", a.config.RateLimit)
	log.Printf("Server address: %s", a.config.ServerAddress)
}

func (a *Agent) runRuntimeCollector(ctx context.Context) {
	pollTicker := time.NewTicker(a.config.PollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			a.collector.CollectRuntime()
		}
	}
}

func (a *Agent) runSystemCollector(ctx context.Context) {
	pollTicker := time.NewTicker(a.config.PollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			if err := a.collector.CollectSystem(); err != nil {
				log.Printf("Failed to collect system metrics: %v", err)
			}
		}
	}
}

func (a *Agent) runReporter(ctx context.Context, jobs chan<- models.Metrics) {
	reportTicker := time.NewTicker(a.config.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-reportTicker.C:
			metrics := a.collector.GetMetricsForReport()
			if len(metrics) == 0 {
				log.Printf("No metrics to send at this time")
				continue
			}

			for _, metric := range metrics {
				select {
				case <-ctx.Done():
					return
				case jobs <- metric:
				}
			}
			log.Printf("Queued %d metrics for sending", len(metrics))
		}
	}
}

func (a *Agent) runSenderWorker(ctx context.Context, workerID int, jobs <-chan models.Metrics) {
	for metric := range jobs {
		if err := a.sender.SendMetricJSONWithContext(ctx, metric); err != nil {
			log.Printf("Worker %d failed to send metric %s: %v", workerID, metric.ID, err)
			continue
		}
		log.Printf("Worker %d sent metric %s", workerID, metric.ID)
	}
}
