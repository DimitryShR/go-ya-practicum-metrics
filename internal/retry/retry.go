package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// DefaultDelays задает интервалы между повторными попытками.
// По умолчанию: 1s, 3s, 5s (три дополнительных попытки).
var DefaultDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

var ErrNilContext = errors.New("nil context")

// Do выполняет операцию с повторами по заданным интервалам.
// Если delays равен nil, используется DefaultDelays.
// shouldRetry определяет, можно ли повторять ошибку.
func Do(ctx context.Context, delays []time.Duration, shouldRetry func(error) bool, op func() error) error {
	if ctx == nil {
		return ErrNilContext
	}
	if delays == nil {
		delays = DefaultDelays
	}

	attempts := len(delays) + 1
	var lastErr error

	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := op()
		if err == nil {
			return nil
		}
		lastErr = err

		if shouldRetry != nil && !shouldRetry(err) {
			return err
		}

		if attempt == len(delays) {
			return fmt.Errorf("retry attempts exceeded after %d tries: %w", attempts, lastErr)
		}

		if !sleep(ctx, delays[attempt]) {
			return ctx.Err()
		}
	}

	return lastErr
}

func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	if ctx.Err() != nil {
		return false
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
