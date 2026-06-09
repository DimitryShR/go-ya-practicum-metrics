package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultRetryCount       = 3
	defaultRetryWaitTime    = 1 * time.Second
	defaultRetryMaxWaitTime = 5 * time.Second
)

// RemoteAuditor — подписчик (наблюдатель), отправляющий аудит-события на удалённый URL.
type RemoteAuditor struct {
	url     string
	ctx     context.Context
	client  *resty.Client
	timeout time.Duration
}

// NewRemoteAuditor создаёт RemoteAuditor с указанным URL, базовым контекстом и таймаутом.
// Использует resty HTTP-клиент с встроенным retry (дефолтные параметры).
func NewRemoteAuditor(ctx context.Context, url string, timeout time.Duration) *RemoteAuditor {
	return NewRemoteAuditorWithRetry(ctx, url, timeout, defaultRetryCount, defaultRetryWaitTime,
		defaultRetryMaxWaitTime)
}

// NewRemoteAuditorWithRetry создаёт RemoteAuditor с кастомными параметрами retry.
func NewRemoteAuditorWithRetry(
	ctx context.Context,
	url string,
	timeout time.Duration,
	retryCount int,
	retryWaitTime time.Duration,
	retryMaxWaitTime time.Duration,
) *RemoteAuditor {
	client := resty.New()
	client.SetRetryCount(retryCount)
	client.SetRetryWaitTime(retryWaitTime)
	client.SetRetryMaxWaitTime(retryMaxWaitTime)
	client.AddRetryCondition(func(resp *resty.Response, err error) bool {
		return shouldRetry(resp, err)
	})

	return &RemoteAuditor{
		url:     url,
		ctx:     ctx,
		client:  client,
		timeout: timeout,
	}
}

// Handle отправляет событие аудита POST-запросом на заданный URL.
// При временных ошибках (сетевые, 5xx) автоматически повторяет запрос.
func (r *RemoteAuditor) Handle(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(r.ctx, r.timeout)
	defer cancel()

	resp, err := r.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(r.url)
	if err != nil {
		return fmt.Errorf("audit remote send: %w", err)
	}

	if resp.StatusCode() >= 300 {
		return fmt.Errorf("audit remote returned status %d", resp.StatusCode())
	}
	return nil
}

// shouldRetry определяет, стоит ли повторять запрос при данной ошибке или статусе.
//
// Ретраим:
//   - Сетевые ошибки: ECONNREFUSED, ECONNRESET, EHOSTUNREACH, ENETUNREACH, ETIMEDOUT
//   - Таймауты: net.Error.Timeout(), context.DeadlineExceeded
//   - HTTP 5xx: временные ошибки сервера
func shouldRetry(resp *resty.Response, err error) bool {
	if err != nil {
		return isRetryableError(err)
	}
	return resp != nil && resp.StatusCode() >= 500
}

// isRetryableError проверяет, является ли ошибка временной.
func isRetryableError(err error) bool {
	// Отмена контекста — не повторяем (shutdown сервера)
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Истечение дедлайна контекста — повторяем
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Сетевые таймауты — повторяем
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Конкретные syscall ошибки — повторяем
	return errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EHOSTUNREACH) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.ETIMEDOUT)
}
