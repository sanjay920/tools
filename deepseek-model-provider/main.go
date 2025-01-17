package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/deepseek-model-provider/validate"

func main() {
	apiKey := os.Getenv("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	cfg := &proxy.Config{
		APIKey:       apiKey,
		Port:         os.Getenv("PORT"),
		UpstreamHost: "api.deepseek.com",
		UseTLS:       true,
		ValidateFn: func(cfg *proxy.Config) error {
			return proxy.DoValidate(cfg, loggerPath, "Invalid DeepSeek Credentials")
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
