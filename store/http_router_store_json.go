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

const jsonExtension = ".json"

func NewHTTPRouterStoreJSON(routerDir string) (*HTTPRouterStoreJSON, error) {
	err := os.MkdirAll(routerDir, os.ModePerm)
	return &HTTPRouterStoreJSON{routerDir: routerDir, prefix: "router_"}, err
}

type HTTPRouterStoreJSON struct {
	routerDir string
	prefix string
}

func (h *HTTPRouterStoreJSON) filepath(name string) string {
	return filepath.Join(h.routerDir, h.prefix + name + jsonExtension)
}

func (h *HTTPRouterStoreJSON) extractName(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, h.prefix), jsonExtension)
}

func (h *HTTPRouterStoreJSON) Delete(ctx context.Context, name string) error {
	return os.Remove(h.filepath(name))
}

func (h *HTTPRouterStoreJSON) GetAll(ctx context.Context, offset, limit int) (map[string]*dynamic.Router, error) {
	ctr := 0

	routers := map[string]*dynamic.Router{}
	// Walk walks the directory in lexical order, so this is fine
	err := filepath.Walk(h.routerDir, func(path string, info os.FileInfo, err error) error {
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
		// decode the content into a dynamic.Router
		router := dynamic.Router{}
		err = json.NewDecoder(f).Decode(&router)
		if err != nil {
			return fmt.Errorf("failed to decode %v: %v", info.Name(), err)
		}
		// add the router to the map
		routers[strings.TrimSuffix(info.Name(), jsonExtension)] = &router

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan %v: %v", h.routerDir, err)
	}

	return routers, nil
}

func (h *HTTPRouterStoreJSON) Get(ctx context.Context, name string) (*dynamic.Router, error) {
	f, err := os.OpenFile(h.filepath(name), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v: %v", name, err)
	}
	defer f.Close()

	router := dynamic.Router{}
	err = json.NewDecoder(f).Decode(&router)
	if err != nil {
		return nil, fmt.Errorf("failed to to decode %v.%v: %v", name, jsonExtension, err)
	}

	return &router, nil
}

func (h *HTTPRouterStoreJSON) Set(ctx context.Context, name string, router *dynamic.Router) error {
	f, err := os.OpenFile(h.filepath(name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open or create %v: %v", name, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(router)
	if err != nil {
		return fmt.Errorf("failed to encode %v.%v: %v", name, jsonExtension, err)
	}

	return nil
}

func (h *HTTPRouterStoreJSON) Names(ctx context.Context, offset, limit int) []string {
	names := make([]string, 0)
	filepath.Walk(h.routerDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasPrefix(info.Name(), h.prefix) {
			return nil
		}
		names = append(names, h.extractName(info.Name()))
		return nil
	})

	return names
}

