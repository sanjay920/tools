package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/vllm-model-provider/validate"

func cleanURL(endpoint string) string {
	return strings.TrimRight(endpoint, "/")
}

func ValidateVLLMAPIKey(cfg *proxy.Config) error {
	endpoint := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT")
	if endpoint == "" {
		const msg = "Invalid vLLM Endpoint"
		log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
		fmt.Printf("{\"error\": \"%s\"}\n", msg)
		return nil
	}

	if cfg.APIKey == "" {
		const msg = "Invalid vLLM Credentials"
		log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
		fmt.Printf("{\"error\": \"%s\"}\n", msg)
		return nil
	}

	endpoint = cleanURL(endpoint)
	return proxy.DoValidate(cfg.APIKey, endpoint+"/v1/models", loggerPath, "Invalid vLLM Configuration")
}

func main() {
	apiKey := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OBOT_VLLM_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	endpoint := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT")
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "OBOT_VLLM_MODEL_PROVIDER_ENDPOINT environment variable not set")
		os.Exit(1)
	}

	endpoint = cleanURL(endpoint)
	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid endpoint URL %q: %v\n", endpoint, err)
		os.Exit(1)
	}

	if u.Scheme == "" {
		if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}

	cfg := &proxy.Config{
		APIKey:          apiKey,
		Port:            os.Getenv("PORT"),
		UpstreamHost:    u.Host,
		UseTLS:          u.Scheme == "https",
		ValidateFn:      ValidateVLLMAPIKey,
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
