package main

import (
	"fmt"
	"log"
	"time"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/deepseek-model-provider/validate"

func ValidateDeepSeekAPIKey(cfg *proxy.Config) error {
	if cfg.APIKey == "" {
		const msg = "Invalid DeepSeek Credentials"
		log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
		fmt.Printf("{\"error\": \"%s\"}\n", msg)
		return nil
	}

	url := "https://api.deepseek.com/v1/models"
	return proxy.DoValidate(cfg.APIKey, url, loggerPath, "Invalid DeepSeek Credentials")
}
