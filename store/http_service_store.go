package store

import (
	"context"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

type HTTPServiceStore interface {
	GetAll(ctx context.Context, offset, limit int) (map[string]*dynamic.Service, error)
	Get(ctx context.Context, name string) (*dynamic.Service, error)
	Set(ctx context.Context, name string, service *dynamic.Service) error
	Delete(ctx context.Context, name string) error
	Names(ctx context.Context, offset, limit int) []string
}

