package handlers

import (
	"context"

	"vault/internal/services"
)

type Handler struct {
	auth  *services.AuthService
	vault *services.VaultService
	pool  *services.WorkerPool
}

func NewHandler(auth *services.AuthService, vault *services.VaultService, pool *services.WorkerPool) *Handler {
	return &Handler{auth: auth, vault: vault, pool: pool}
}

func (h *Handler) runInPool(ctx context.Context, job func() (any, error)) (any, error) {
	resultCh := make(chan any, 1)
	errCh := make(chan error, 1)

	h.pool.Submit(func() {
		res, err := job()
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case res := <-resultCh:
		return res, nil
	}
}
