package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"os"
	"path/filepath"
	"strings"
)

func NewHTTPMiddlewareStoreJSON(middlewareDir string) (*HTTPMiddlewareStoreJSON, error) {
	err := os.MkdirAll(middlewareDir, os.ModePerm)
	return &HTTPMiddlewareStoreJSON{middlewareDir: middlewareDir, prefix: "middleware_"}, err
}

type HTTPMiddlewareStoreJSON struct {
	middlewareDir string
	prefix string
}

func (h *HTTPMiddlewareStoreJSON) filepath(name string) string {
	return filepath.Join(h.middlewareDir, h.prefix + name + jsonExtension)
}

func (h *HTTPMiddlewareStoreJSON) extractName(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, h.prefix), jsonExtension)
}

func (h *HTTPMiddlewareStoreJSON) Delete(ctx context.Context, name string) error {
	return os.Remove(h.filepath(name))
}

func (h *HTTPMiddlewareStoreJSON) GetAll(ctx context.Context, offset, limit int) (map[string]*dynamic.Middleware, error) {
	ctr := 0

	middlewares := map[string]*dynamic.Middleware{}
	// Walk walks the directory in lexical order, so this is fine
	err := filepath.Walk(h.middlewareDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasPrefix(info.Name(), h.prefix) {
			return nil
		}
		// skip
		if ctr < offset {
			ctr++
			return nil
		}
		if limit == 0 {
			return nil
		}
		// decrease the limit to keep track of the state
		limit--
		f, err := os.OpenFile(path, os.O_RDONLY, info.Mode())
		if err != nil {
			return fmt.Errorf("failed to open %v: %v", path, err)
		}
		defer f.Close()
		// decode the content into a dynamic.Middleware
		middleware := dynamic.Middleware{}
		err = json.NewDecoder(f).Decode(&middleware)
		if err != nil {
			return fmt.Errorf("failed to decode %v: %v", info.Name(), err)
		}
		// add the middleware to the map
		middlewares[h.extractName(info.Name())] = &middleware

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan %v: %v", h.middlewareDir, err)
	}

	return middlewares, nil
}

func (h *HTTPMiddlewareStoreJSON) Get(ctx context.Context, name string) (*dynamic.Middleware, error) {
	f, err := os.OpenFile(h.filepath(name), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v: %v", name, err)
	}
	defer f.Close()

	middleware := dynamic.Middleware{}
	err = json.NewDecoder(f).Decode(&middleware)
	if err != nil {
		return nil, fmt.Errorf("failed to to decode %v.%v: %v", name, jsonExtension, err)
	}

	return &middleware, nil
}

func (h *HTTPMiddlewareStoreJSON) Set(ctx context.Context, name string, middleware *dynamic.Middleware) error {
	f, err := os.OpenFile(h.filepath(name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open or create %v: %v", name, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(middleware)
	if err != nil {
		return fmt.Errorf("failed to encode %v.%v: %v", name, jsonExtension, err)
	}

	return nil
}

func (h *HTTPMiddlewareStoreJSON) Names(ctx context.Context, offset, limit int) []string {
	names := make([]string, 0)
	filepath.Walk(h.middlewareDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasPrefix(info.Name(), h.prefix) {
			return nil
		}
		names = append(names, h.extractName(info.Name()))
		return nil
	})

	return names
}