package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

type Config struct {
	APIKey          string
	Port            string
	UpstreamHost    string
	UseTLS          bool
	ValidateFn      func(cfg *Config) error
	RewriteModelsFn func(*http.Response) error
	PathPrefix      string
}

type server struct {
	cfg *Config
}

func Run(cfg *Config) error {
	if cfg.Port == "" {
		cfg.Port = "8000"
	}
	if cfg.UpstreamHost == "" {
		cfg.UpstreamHost = "api.openai.com"
		cfg.UseTLS = true
	}

	if cfg.RewriteModelsFn == nil {
		cfg.RewriteModelsFn = DefaultRewriteModelsResponse
	}

	if cfg.ValidateFn != nil {
		if err := cfg.ValidateFn(cfg); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	s := &server{cfg: cfg}

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", s.healthz)
	mux.Handle("/v1/models", &httputil.ReverseProxy{
		Director:       s.proxyDirector,
		ModifyResponse: cfg.RewriteModelsFn,
	})
	mux.Handle("/v1/", &httputil.ReverseProxy{
		Director: s.proxyDirector,
	})

	httpServer := &http.Server{
		Addr:    "127.0.0.1:" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("Starting proxy on port %s → host=%s\n", cfg.Port, cfg.UpstreamHost)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("http://127.0.0.1:" + s.cfg.Port))
}

func (s *server) proxyDirector(req *http.Request) {
	scheme := "https"
	if !s.cfg.UseTLS {
		scheme = "http"
	}
	req.URL.Scheme = scheme
	req.URL.Host = s.cfg.UpstreamHost
	req.Host = req.URL.Host

	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	if s.cfg.PathPrefix != "" {
		if !strings.HasPrefix(req.URL.Path, s.cfg.PathPrefix) {
			req.URL.Path = s.cfg.PathPrefix + req.URL.Path
		}
	}
}

func Validate(cfg *Config) error {
	if cfg.ValidateFn == nil {
		return nil
	}
	return cfg.ValidateFn(cfg)
}
