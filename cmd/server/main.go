package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/audit"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/config"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/handler"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/repository"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/service"
	"github.com/DimitryShR/go-ya-practicum-metrics/internal/sign"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

type serverEntry struct {
	srv     *http.Server
	timeout time.Duration
}

const (
	mainShutdownTimeout  = 30 * time.Second
	pprofShutdownTimeout = 3 * time.Second
)

type storageMode string

const (
	modeMemory   storageMode = "memory"
	modeFile     storageMode = "file"
	modePostgres storageMode = "postgres"
)

func main() {
	printBuildInfo()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var storage service.Storage
	var db *sql.DB

	cfg := config.NewServerConfig()

	fmt.Printf("Starting server on %s\n", cfg.Address)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return err
	}
	defer logger.Log.Sync()

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	var signer *sign.Signer
	if cfg.SignKey != "" {
		signer = sign.NewSigner(cfg.SignKey)
	}

	mode := getMode(cfg)
	logger.Log.Info("Storage mode selected", zap.String("mode", string(mode)))

	// Контекст для периодической записи метрик в файл.
	// Отменяется при shutdown, чтобы корректно остановить фоновую горутину.
	storageCtx, storageCancel := context.WithCancel(context.Background())
	defer storageCancel()

	switch mode {
	case modePostgres:
		pgStorage, err := initPostgresStorage(cfg, &db)
		if err != nil {
			return fmt.Errorf("initialize PostgreSQL storage: %w", err)
		}
		if db != nil {
			defer db.Close()
		}
		storage = pgStorage
	case modeFile:
		fileStorage, err := initFileStorage(storageCtx, cfg)
		if err != nil {
			return fmt.Errorf("initialize file storage: %w", err)
		}
		storage = fileStorage
	case modeMemory:
		storage = repository.NewMemStorage()
	}

	metricService := service.NewMetricService(storage)

	auditCtx, auditCancel := context.WithCancel(context.Background())

	auditPublisher := audit.NewAuditPublisher()

	if cfg.AuditFile != "" {
		fileAuditor, err := audit.NewFileAuditor(auditCtx, cfg.AuditFile)
		if err != nil {
			defer auditCancel()
			defer auditPublisher.Shutdown()
			return fmt.Errorf("initialize file auditor: %w", err)
		}
		// Дожидаемся сохранения всех данных в файл
		defer fileAuditor.Wait()
		auditPublisher.Register(fileAuditor)
		logger.Log.Info("Audit file sink enabled", zap.String("path", cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditPublisher.Register(audit.NewRemoteAuditor(auditCtx, cfg.AuditURL, 10*time.Second))
		logger.Log.Info("Audit remote sink enabled", zap.String("url", cfg.AuditURL))
	}

	// Отменяем контекст auditCtx, который обеспечивает корректное завершение аудиторов
	defer auditCancel()
	// Останавливаем рассылку событий и после отменяем контекст auditCtx
	defer auditPublisher.Shutdown()

	metricHandler := handler.NewMetricHandler(metricService, auditPublisher)

	pingHandler := handler.NewPingHandler(db)
	router := newRouter(metricHandler, pingHandler, signer)

	mainSrv := runServer(cfg.Address, router)

	var pprofSrv *http.Server
	if cfg.PprofAddress != "" {
		pprofSrv = runPprofServer(cfg.PprofAddress)
	}

	return shutdownServer(
		serverEntry{mainSrv, mainShutdownTimeout},
		serverEntry{pprofSrv, pprofShutdownTimeout},
	)
}

// runServer запускает HTTP-сервер с таймаутами.
func runServer(addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в горутине. ListenAndServe возвращает http.ErrServerClosed при вызове Shutdown.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	logger.Log.Info("Server started", zap.String("address", addr))
	return srv
}

// runPprofServer запускает HTTP-сервер для Pprof.
func runPprofServer(addr string) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux,
	}

	// Запуск сервера в горутине. ListenAndServe возвращает http.ErrServerClosed при вызове Shutdown.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Pprof server failed", zap.Error(err))
		}
	}()
	logger.Log.Info("Starting pprof server", zap.String("address", addr))
	return srv
}

// shutdownServer обеспечивает graceful shutdown при получении сигналов SIGINT или SIGTERM.
func shutdownServer(servers ...serverEntry) error {
	// Ожидание сигнала завершения (SIGINT или SIGTERM)
	sigCtx, sigStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigStop()
	<-sigCtx.Done()
	logger.Log.Info("Received signal, shutting down server...")

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, entry := range servers {
		if entry.srv == nil {
			continue
		}
		wg.Add(1)
		go func(e serverEntry) {
			defer wg.Done()
			// Параметризированный таймаут на graceful завершение сервера
			ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
			defer cancel()
			if err := e.srv.Shutdown(ctx); err != nil {
				logger.Log.Error("Server forced to shutdown", zap.String("server address", e.srv.Addr), zap.Error(err))
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(entry)
	}
	wg.Wait()

	if len(errs) == 0 {
		logger.Log.Info("Server stopped gracefully")
	}
	return errors.Join(errs...)
}

func newRouter(
	metricHandler *handler.MetricHandler,
	pingHandler *handler.PingHandler,
	signer *sign.Signer,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.LogRequest, middleware.GzipMiddleware)
	if signer != nil {
		r.Use(middleware.SignMiddleware(signer))
	}

	r.Post("/update", metricHandler.UpdateMetricHandlerJSON)
	r.With(middleware.ParseUpdatePathHandler).Post("/update/*", metricHandler.UpdateMetricHandler)

	r.Post("/updates", metricHandler.UpdateMetricsHandlerJSON)
	r.Post("/value", metricHandler.GetMetricValueJSON)
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetricValue)
	r.Get("/", metricHandler.GetAllMetrics)
	r.Get("/ping", pingHandler.Ping)

	return r
}

func getMode(cfg *config.ServerConfig) storageMode {
	dbDSN := cfg.DBDsn.GetKeywordDSN()
	if dbDSN != "" {
		return modePostgres
	}
	if cfg.FileStoragePath != "" {
		return modeFile
	}
	return modeMemory
}

func initPostgresStorage(cfg *config.ServerConfig, db **sql.DB) (*repository.PgStorage, error) {
	migrateURL := cfg.DBMigrateDsn.GetURL()
	if migrateURL == "" {
		logger.Log.Info("DB migrate DSN not provided, try using main DB DSN for migrations")
		migrateURL = cfg.DBDsn.GetURL()
	}
	if err := runMigrations(migrateURL); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	logger.Log.Info("Database migrations completed successfully")

	dbDSN := cfg.DBDsn.GetKeywordDSN()
	newDB, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := newDB.PingContext(ctx); err != nil {
		_ = newDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	*db = newDB
	return repository.NewPgStorage(newDB), nil
}

func runMigrations(migrateDSN string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory for migrations: %w", err)
	}
	sourceURL := "file://" + filepath.Join(wd, "migrations")

	m, err := migrate.New(sourceURL, migrateDSN)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func initFileStorage(ctx context.Context, cfg *config.ServerConfig) (*repository.MemStorage, error) {
	memStorage := repository.NewMemStorage()

	if cfg.Restore {
		if ok, err := memStorage.LoadFromFile(ctx, cfg.FileStoragePath); err != nil {
			return nil, fmt.Errorf("restore metrics from file: %w", err)
		} else if ok {
			logger.Log.Info("Metrics restored from file", zap.String("file", cfg.FileStoragePath))
		}
	}

	switch {
	case cfg.StoreInterval < 0:
		return nil, fmt.Errorf("store interval must be non-negative, got %s", cfg.StoreInterval)
	case cfg.StoreInterval == 0:
		memStorage.EnableSyncSave(cfg.FileStoragePath)
		logger.Log.Info("Enabled synchronous metrics persistence", zap.String("file", cfg.FileStoragePath))
	default:
		go func(ctx context.Context, path string, interval time.Duration) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := memStorage.SaveToFile(ctx, path); err != nil {
						logger.Log.Error("Cannot save metrics to file", zap.Error(err), zap.String("file", path))
					}
				case <-ctx.Done():
					return
				}
			}
		}(ctx, cfg.FileStoragePath, cfg.StoreInterval)
		logger.Log.Info(
			"Enabled periodic metrics persistence",
			zap.Duration("interval", cfg.StoreInterval),
			zap.String("file", cfg.FileStoragePath),
		)
	}

	return memStorage, nil
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
