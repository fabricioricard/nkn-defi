package repo

import (
    "context"
    "github.com/fabricioricard/nkn-defi/internal/domain"
)

type PoolRepository interface {
    Create(ctx context.Context, pool *domain.Pool) error
    GetByID(ctx context.Context, id string) (*domain.Pool, error)
    List(ctx context.Context) ([]*domain.Pool, error)
    Update(ctx context.Context, pool *domain.Pool) error
}