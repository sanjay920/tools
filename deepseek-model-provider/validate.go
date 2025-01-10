package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/deepseek-model-provider/validate"

func ValidateDeepSeekAPIKey(cfg *proxy.Config) error {
	if cfg.APIKey == "" {
		fmt.Printf("{\"error\": \"%s\"}\n", "Invalid DeepSeek Credentials")
		os.Exit(0)
	}

	url := "https://api.deepseek.com/v1/models"
	return proxy.DoValidate(cfg.APIKey, url, loggerPath, "Invalid DeepSeek Credentials")
}
