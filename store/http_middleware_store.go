package store

import (
	"context"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

type HTTPMiddlewareStore interface {
	GetAll(ctx context.Context, offset, limit int) (map[string]*dynamic.Middleware, error)
	Get(ctx context.Context, name string) (*dynamic.Middleware, error)
	Set(ctx context.Context, name string, middleware *dynamic.Middleware) error
	Delete(ctx context.Context, name string) error
	Names(ctx context.Context, offset, limit int) []string
}

