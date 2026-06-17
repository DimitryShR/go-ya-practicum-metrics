package repository

import (
	"context"
	"errors"
)

// ErrNilContext возвращается, если контекст равен nil.
var ErrNilContext = errors.New("nil context")

// requireContext проверяет, что контекст не равен nil.
// Возвращает ErrNilContext если ctx == nil, иначе возвращает ctx.Err().
func requireContext(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	return ctx.Err()
}
