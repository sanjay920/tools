package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/xai-model-provider/validate"

func ValidateXAIAPIKey(cfg *proxy.Config) error {
	if cfg.APIKey == "" {
		const msg = "Invalid xAI Credentials"
		log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
		fmt.Printf("{\"error\": \"%s\"}\n", msg)
		return nil
	}

	url := "https://api.x.ai/v1/models"
	return proxy.DoValidate(cfg.APIKey, url, loggerPath, "Invalid xAI Credentials")
}

// RewriteGrokModels marks only Grok models as LLMs
func RewriteGrokModels(resp *http.Response) error {
	rewriteFn := proxy.RewriteAllModelsWithUsage("llm", func(modelID string) bool {
		return strings.HasPrefix(modelID, "grok-")
	})
	return rewriteFn(resp)
}

func main() {
	apiKey := os.Getenv("OBOT_XAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OBOT_XAI_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	cfg := &proxy.Config{
		APIKey:          apiKey,
		Port:            os.Getenv("PORT"),
		UpstreamHost:    "api.x.ai",
		UseTLS:          true,
		ValidateFn:      ValidateXAIAPIKey,
		RewriteModelsFn: RewriteGrokModels,
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
