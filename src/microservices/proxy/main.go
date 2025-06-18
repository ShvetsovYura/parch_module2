package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

type RouteItem struct {
	Name   string `yaml:"name"`
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
	Strip  bool   `yaml:"strip"`
}

type ProxyConfig struct {
	Routes []RouteItem `yaml:"routes"`
}

func findPath(cfg ProxyConfig, path string) (bool, *RouteItem) {
	for _, r := range cfg.Routes {
		if r.Path == path {
			return true, &r
		}
	}
	return false, nil
}

func NewProxy(target *url.URL) *httputil.ReverseProxy {

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(r *http.Request) {
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Host = target.Host
	}
	return proxy
}

func getTarget(cfg ProxyConfig, path string, is_use_microservices bool, trafficPercent float64) (*url.URL, error) {
	monolithUrl, err := url.Parse(os.Getenv("MONOLITH_URL"))
	if !is_use_microservices {
		return monolithUrl, err
	}
	randomValue := rand.Float64()
	if randomValue > trafficPercent {
		return monolithUrl, err
	}

	ok, routePath := findPath(cfg, path)
	if !ok {
		return monolithUrl, err

	}
	return url.Parse(routePath.Target)
}

func main() {
	migraion_env := os.Getenv("MOVIES_MIGRATION_PERCENT")
	port_env := os.Getenv("PORT")
	is_gradual_migration := os.Getenv("GRADUAL_MIGRATION") == "true"

	if migraion_env == "" {
		migraion_env = "0"
	}
	if port_env == "" {
		port_env = "8080"
	}

	port := ":" + port_env

	slog.Info("Starting proxy service",
		slog.String("migration percent", migraion_env),
		slog.Bool("migration enabled", is_gradual_migration),
		slog.String("webapi port", port),
	)
	trafficSplit, _ := strconv.ParseFloat(migraion_env, 64)
	trafficSplit = trafficSplit / 100.0
	routesConfig := ProxyConfig{
		Routes: []RouteItem{
			{
				Name:   "Movie service",
				Path:   "/api/movies",
				Target: os.Getenv("MOVIES_SERVICE_URL"),
				Strip:  false,
			}, {
				Name:   "Events service",
				Path:   "/api/events",
				Target: os.Getenv("EVENTS_SERVICE_URL"),
				Strip:  false,
			},
		},
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "only GET", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"status": true})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		target, err := getTarget(routesConfig, req.URL.Path, is_gradual_migration, trafficSplit)
		if err != nil {
			log.Fatal(err.Error())
		}
		slog.Info("Received request", slog.String("path", req.URL.Path), slog.String("to", target.Host))
		proxy := NewProxy(target)
		proxy.ServeHTTP(w, req)
	})
	log.Fatal(http.ListenAndServe(port, nil))
}
