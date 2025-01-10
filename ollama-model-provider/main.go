package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/ollama-model-provider/validate"

func cleanHost(host string) string {
	return strings.TrimRight(host, "/")
}

func ValidateOllamaAPIKey(cfg *proxy.Config) error {
	host := os.Getenv("OBOT_OLLAMA_MODEL_PROVIDER_HOST")
	if host == "" {
		const msg = "Invalid Ollama Host"
		log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
		fmt.Printf("{\"error\": \"%s\"}\n", msg)
		return nil
	}

	host = cleanHost(host)
	url := fmt.Sprintf("http://%s/v1/models", host)
	return proxy.DoValidate("", url, loggerPath, "Invalid Ollama Host")
}

func main() {
	host := os.Getenv("OBOT_OLLAMA_MODEL_PROVIDER_HOST")
	if host == "" {
		host = "127.0.0.1:11434"
	}
	host = cleanHost(host)

	cfg := &proxy.Config{
		APIKey:          "",
		Port:            os.Getenv("PORT"),
		UpstreamHost:    host,
		UseTLS:          false,
		ValidateFn:      ValidateOllamaAPIKey,
		RewriteModelsFn: proxy.RewriteAllModelsWithUsage("llm"),
		PathPrefix:      "/v1",
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
