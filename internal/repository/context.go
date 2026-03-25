package repository

import (
	"context"
	"errors"
)

var ErrNilContext = errors.New("nil context")

func requireContext(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	return ctx.Err()
}
