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

func NewHTTPServiceStoreJSON(serviceDir string) (*HTTPServiceStoreJSON, error) {
	err := os.MkdirAll(serviceDir, os.ModePerm)
	return &HTTPServiceStoreJSON{serviceDir: serviceDir, prefix: "service_"}, err
}

type HTTPServiceStoreJSON struct {
	serviceDir string
	prefix string
}

func (h *HTTPServiceStoreJSON) filepath(name string) string {
	return filepath.Join(h.serviceDir, h.prefix + name + jsonExtension)
}

func (h *HTTPServiceStoreJSON) extractName(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, h.prefix), jsonExtension)
}

func (h *HTTPServiceStoreJSON) Delete(ctx context.Context, name string) error {
	return os.Remove(h.filepath(name))
}

func (h *HTTPServiceStoreJSON) GetAll(ctx context.Context, offset, limit int) (map[string]*dynamic.Service, error) {
	ctr := 0

	services := map[string]*dynamic.Service{}
	// Walk walks the directory in lexical order, so this is fine
	err := filepath.Walk(h.serviceDir, func(path string, info os.FileInfo, err error) error {
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
		// decode the content into a dynamic.Service
		service := dynamic.Service{}
		err = json.NewDecoder(f).Decode(&service)
		if err != nil {
			return fmt.Errorf("failed to decode %v: %v", info.Name(), err)
		}
		// add the service to the map
		services[strings.TrimSuffix(info.Name(), jsonExtension)] = &service

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan %v: %v", h.serviceDir, err)
	}

	return services, nil
}

func (h *HTTPServiceStoreJSON) Get(ctx context.Context, name string) (*dynamic.Service, error) {
	f, err := os.OpenFile(h.filepath(name), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v: %v", name, err)
	}
	defer f.Close()

	service := dynamic.Service{}
	err = json.NewDecoder(f).Decode(&service)
	if err != nil {
		return nil, fmt.Errorf("failed to to decode %v.%v: %v", name, jsonExtension, err)
	}

	return &service, nil
}

func (h *HTTPServiceStoreJSON) Set(ctx context.Context, name string, service *dynamic.Service) error {
	f, err := os.OpenFile(h.filepath(name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open or create %v: %v", name, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(service)
	if err != nil {
		return fmt.Errorf("failed to encode %v.%v: %v", name, jsonExtension, err)
	}

	return nil
}

func (h *HTTPServiceStoreJSON) Names(ctx context.Context, offset, limit int) []string {
	names := make([]string, 0)
	filepath.Walk(h.serviceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasPrefix(info.Name(), h.prefix) {
			return nil
		}
		names = append(names, h.extractName(info.Name()))
		return nil
	})

	return names
}


