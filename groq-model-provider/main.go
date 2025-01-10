package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

const loggerPath = "/tools/groq-model-provider/validate"

func ValidateGroqAPIKey(cfg *proxy.Config) error {
	if cfg.APIKey == "" {
		const msg = "Invalid Groq Credentials"
		log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
		fmt.Printf("{\"error\": \"%s\"}\n", msg)
		return nil
	}

	url := "https://api.groq.com/openai/v1/models"
	return proxy.DoValidate(cfg.APIKey, url, loggerPath, "Invalid Groq Credentials")
}

func main() {
	apiKey := os.Getenv("OBOT_GROQ_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OBOT_GROQ_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	cfg := &proxy.Config{
		APIKey:          apiKey,
		Port:            os.Getenv("PORT"),
		UpstreamHost:    "api.groq.com",
		UseTLS:          true,
		ValidateFn:      ValidateGroqAPIKey,
		RewriteModelsFn: proxy.RewriteAllModelsWithUsage("llm"),
		PathPrefix:      "/openai",
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
