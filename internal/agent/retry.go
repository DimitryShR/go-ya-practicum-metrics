package agent

import (
	"context"
	"errors"
	"net"
	"syscall"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/retry"
)

// withRetry выполняет операцию с retry для временных сетевых ошибок.
func (c *MetricsClient) withRetry(ctx context.Context, op func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return retry.Do(ctx, nil, isRetryableRequestError, op)
}

// isRetryableRequestError проверяет, является ли ошибка временной и стоит ли повторять запрос.
// Повторяет: context.DeadlineExceeded, сетевые таймауты, syscall ошибки соединения.
// Не повторяет: context.Canceled.
func isRetryableRequestError(err error) bool {
	if err == nil {
		return false
	}

	// Если контекст был отменен -> не повторять
	if errors.Is(err, context.Canceled) {
		return false
	}
	// Если истек таймаут контекста -> повторить
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	// Сетевой таймаут -> повторить
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	switch {
	case
		errors.Is(err, syscall.ECONNREFUSED), // отказ в соединении
		errors.Is(err, syscall.ECONNRESET),   // соединение сброшено
		errors.Is(err, syscall.EHOSTUNREACH), // хост недоступен
		errors.Is(err, syscall.ENETUNREACH),  // сеть недоступна
		errors.Is(err, syscall.ETIMEDOUT):    // истек таймаут
		return true
	}

	return false
}
