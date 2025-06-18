package main

import (
	"encoding/json"
	"log"
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

func NewProxy(target *url.URL, cfg ProxyConfig) *httputil.ReverseProxy {

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(r *http.Request) {
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Host = target.Host
	}
	return proxy
}

func getTarget(cfg ProxyConfig, path string, is_use_microservices bool, trafficPercent float64) (*url.URL, error) {
	if !is_use_microservices {
		return url.Parse(os.Getenv("MONOLITH_URL"))
	}
	randomValue := rand.Float64()
	if randomValue > trafficPercent {
		return url.Parse(os.Getenv("MONOLITH_URL"))
	}

	ok, routePath := findPath(cfg, path)
	if !ok {
		return url.Parse(os.Getenv("MONOLITH_URL"))

	}
	return url.Parse(routePath.Target)
}

func main() {
	trafficSplit, _ := strconv.ParseFloat(os.Getenv("MOVIES_MIGRATION_PERCENT"), 64)
	cfg := ProxyConfig{
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
		target, err := getTarget(cfg, req.URL.Path, os.Getenv("GRADUAL_MIGRATION") == "true", trafficSplit)
		if err != nil {
			log.Fatal(err.Error())
		}

		proxy := NewProxy(target, cfg)
		proxy.ServeHTTP(w, req)
	})
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
