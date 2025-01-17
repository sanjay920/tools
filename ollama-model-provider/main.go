package main

import (
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/ollama-model-provider/validate"

func cleanHost(host string) string {
	return strings.TrimRight(host, "/")
}

func main() {
	host := os.Getenv("OBOT_OLLAMA_MODEL_PROVIDER_HOST")
	if host == "" {
		host = "127.0.0.1:11434"
	}
	host = cleanHost(host)

	cfg := &proxy.Config{
		APIKey:       "",
		Port:         os.Getenv("PORT"),
		UpstreamHost: host,
		UseTLS:       false,
		ValidateFn: func(cfg *proxy.Config) error {
			return proxy.DoValidate(cfg, loggerPath, "Invalid Ollama Host")
		},
		RewriteModelsFn: proxy.RewriteAllModelsWithUsage("llm"),
	}

	if cfg.Port == "" {
		cfg.Port = "8000"
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := proxy.Validate(cfg); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
