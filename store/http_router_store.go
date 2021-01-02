package store

import (
	"context"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

type HTTPRouterStore interface {
	GetAll(ctx context.Context, offset, limit int) (map[string]*dynamic.Router, error)
	Get(ctx context.Context, name string) (*dynamic.Router, error)
	Set(ctx context.Context, name string, router *dynamic.Router) error
	Delete(ctx context.Context, name string) error
	Names(ctx context.Context, offset, limit int) []string
}
