package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"kommandeur/store"
	"net/http"
	"time"
)

func main() {
	// httpRouterStore, err := store.NewJsonStore(store.ModeJson)
	var httpRouterStore store.HTTPRouterStore
	var err error

	httpRouterStore, err = store.NewHTTPRouterStoreJSON("routers")
	if err != nil {
		fmt.Printf("failed to create a new httprouterstore: %v", err)
		return
	}

	var httpServiceStore store.HTTPServiceStore
	httpServiceStore, err = store.NewHTTPServiceStoreJSON("services")
	if err != nil {
		fmt.Printf("failed to create a new httpservicestore: %v", err)
		return
	}

	var httpMiddlewareStore store.HTTPMiddlewareStore
	httpMiddlewareStore, err = store.NewHTTPMiddlewareStoreJSON("middlewares")
	if err != nil {
		fmt.Printf("failed to create a new httpmiddlewarestore: %v", err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		routers, err := httpRouterStore.GetAll(ctx, 0, -1)
		if err != nil {
			fmt.Printf("failed to get routers from store: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		services, err := httpServiceStore.GetAll(ctx, 0, -1)
		if err != nil {
			fmt.Printf("failed to get services from store: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middlewares, err := httpMiddlewareStore.GetAll(ctx, 0, -1)
		if err != nil {
			fmt.Printf("failed to get middlewares from store: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		conf := dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers:     routers,
				Services:    services,
				Middlewares: middlewares,
				Models:      nil,
			},
			TCP:  nil,
			UDP:  nil,
			TLS:  nil,
		}

		switch contentType {
		case "toml":
			w.Header().Set("Content-Type", "application/toml")
			toml.NewEncoder(w).Encode(conf)
		case "json":
			fallthrough
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(conf)
		}
	})

	v1Router := r.PathPrefix("/v1").Subrouter()
	httpRouter := v1Router.PathPrefix("/http").Subrouter()
	httpRouter.HandleFunc("/routers", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		names := httpRouterStore.Names(ctx ,0, 20)

		type HML map[string]struct{
			Href string `json:"href"`
		}

		type Name struct{
			Name string `json:"name"`
			Links HML `json:"_links"`
		}

		type Response struct {
			Names []Name `json:"names"`
			Links HML `json:"_links"`
		}

		response := Response{
			Names: make([]Name, 0),
			Links: HML{
				"self": {
					Href: "/v1/http/routers",
				},
			},
		}

		for _, name := range names {
			response.Names = append(response.Names, Name{
				Name:  name,
				Links: HML{
					"self": {
						Href: "/v1/http/router/" + name,
					},
				},
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods(http.MethodGet)
	httpRouter.HandleFunc("/router", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		configuration := dynamic.Configuration{}
		switch contentType {
		case "toml":
			_, err := toml.DecodeReader(r.Body, &configuration)
			if err != nil {
				fmt.Printf("failed to add a new router from toml: failed to decode r.Body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "json":
			fallthrough
		default:
			err := json.NewDecoder(r.Body).Decode(&configuration)
			if err != nil {
				fmt.Printf("failed to add a new router from json: failed to decode r.Body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		for name, router := range configuration.HTTP.Routers {
			err = httpRouterStore.Set(ctx, name, router)
			if err != nil {
				fmt.Printf("failed store the router in httpRouterStore: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)
	httpRouter.HandleFunc("/router/{name:[a-zA-Z0-9=\\-\\/]+}", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		name, ok := mux.Vars(r)["name"]
		if !ok {
			fmt.Printf("did not find name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		router, err := httpRouterStore.Get(ctx, name)
		if err != nil {
			fmt.Printf("failed to get router for %v from store: %v\n", name, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch contentType {
		case "toml":
			w.Header().Set("Content-Type", "application/toml")
			toml.NewEncoder(w).Encode(router)
		case "json":
			fallthrough
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(router)
		}
	}).Methods(http.MethodGet)
	httpRouter.HandleFunc("/router/{name:[a-zA-Z0-9=\\-\\/]+}", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		name, ok := mux.Vars(r)["name"]
		if !ok {
			fmt.Printf("did not find name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := httpRouterStore.Delete(ctx, name)
		if err != nil {
			fmt.Printf("could not delete %v: %v", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	httpRouter.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		names := httpServiceStore.Names(ctx, 0, -1)

		type HML map[string]struct{
			Href string `json:"href"`
		}

		type Name struct{
			Name string `json:"name"`
			Links HML `json:"_links"`
		}

		type Response struct {
			Names []Name `json:"names"`
			Links HML `json:"_links"`
		}

		response := Response{
			Names: make([]Name, 0),
			Links: HML{
				"self": {
					Href: "/v1/http/services",
				},
			},
		}

		for _, name := range names {
			response.Names = append(response.Names, Name{
				Name:  name,
				Links: HML{
					"self": {
						Href: "/v1/http/service/" + name,
					},
				},
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods(http.MethodGet)
	httpRouter.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		configuration := dynamic.Configuration{}
		switch contentType {
		case "toml":
			_, err := toml.DecodeReader(r.Body, &configuration)
			if err != nil {
				fmt.Printf("failed to add a new service from toml: failed to decode r.Body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "json":
			fallthrough
		default:
			err := json.NewDecoder(r.Body).Decode(&configuration)
			if err != nil {
				fmt.Printf("failed to add a new service from json: failed to decode r.Body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		for name, service := range configuration.HTTP.Services {
			err = httpServiceStore.Set(ctx, name, service)
			if err != nil {
				fmt.Printf("failed store the service in httpServiceStore: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)
	httpRouter.HandleFunc("/service/{name:[a-zA-Z0-9=\\-\\/]+}", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		name, ok := mux.Vars(r)["name"]
		if !ok {
			fmt.Printf("did not find name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		service, err := httpServiceStore.Get(ctx, name)
		if err != nil {
			fmt.Printf("failed to get service for %v from store: %v\n", name, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch contentType {
		case "toml":
			w.Header().Set("Content-Type", "application/toml")
			toml.NewEncoder(w).Encode(service)
		case "json":
			fallthrough
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(service)
		}
	}).Methods(http.MethodGet)
	httpRouter.HandleFunc("/service/{name:[a-zA-Z0-9=\\-\\/]+}", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		name, ok := mux.Vars(r)["name"]
		if !ok {
			fmt.Printf("did not find name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := httpServiceStore.Delete(ctx, name)
		if err != nil {
			fmt.Printf("could not delete %v: %v", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	httpRouter.HandleFunc("/middlewares", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		names := httpMiddlewareStore.Names(ctx, 0, -1)

		type HML map[string]struct{
			Href string `json:"href"`
		}

		type Name struct{
			Name string `json:"name"`
			Links HML `json:"_links"`
		}

		type Response struct {
			Names []Name `json:"names"`
			Links HML `json:"_links"`
		}

		response := Response{
			Names: make([]Name, 0),
			Links: HML{
				"self": {
					Href: "/v1/http/middlewares",
				},
			},
		}

		for _, name := range names {
			response.Names = append(response.Names, Name{
				Name:  name,
				Links: HML{
					"self": {
						Href: "/v1/http/middleware/" + name,
					},
				},
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods(http.MethodGet)
	httpRouter.HandleFunc("/middleware", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		configuration := dynamic.Configuration{}
		switch contentType {
		case "toml":
			_, err := toml.DecodeReader(r.Body, &configuration)
			if err != nil {
				fmt.Printf("failed to add a new middleware from toml: failed to decode r.Body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "json":
			fallthrough
		default:
			err := json.NewDecoder(r.Body).Decode(&configuration)
			if err != nil {
				fmt.Printf("failed to add a new middleware from json: failed to decode r.Body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		for name, middleware := range configuration.HTTP.Middlewares {
			err = httpMiddlewareStore.Set(ctx, name, middleware)
			if err != nil {
				fmt.Printf("failed store the middleware in httpMiddlewareStore: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)
	httpRouter.HandleFunc("/middleware/{name:[a-zA-Z0-9=\\-\\/]+}", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		name, ok := mux.Vars(r)["name"]
		if !ok {
			fmt.Printf("did not find name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		v := r.URL.Query()
		contentType := v.Get("type")
		if contentType == "" {
			contentType = "json"
		}

		middleware, err := httpMiddlewareStore.Get(ctx, name)
		if err != nil {
			fmt.Printf("failed to get middleware for %v from store: %v\n", name, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch contentType {
		case "toml":
			w.Header().Set("Content-Type", "application/toml")
			toml.NewEncoder(w).Encode(middleware)
		case "json":
			fallthrough
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(middleware)
		}
	}).Methods(http.MethodGet)
	httpRouter.HandleFunc("/middleware/{name:[a-zA-Z0-9=\\-\\/]+}", func(w http.ResponseWriter, r *http.Request) {
		ctx,cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()

		name, ok := mux.Vars(r)["name"]
		if !ok {
			fmt.Printf("did not find name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := httpMiddlewareStore.Delete(ctx, name)
		if err != nil {
			fmt.Printf("could not delete %v: %v", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	http.ListenAndServe(":8080", r)
}