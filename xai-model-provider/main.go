package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/xai-model-provider/validate"

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
		APIKey:       apiKey,
		Port:         os.Getenv("PORT"),
		UpstreamHost: "api.x.ai",
		UseTLS:       true,
		ValidateFn: func(cfg *proxy.Config) error {
			return proxy.DoValidate(cfg, loggerPath, "Invalid xAI Credentials")
		},
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
