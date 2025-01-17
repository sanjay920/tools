package proxy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func handleValidationError(loggerPath, msg string) error {
	slog.Error(msg, "logger", loggerPath)
	fmt.Printf("{\"error\": \"%s\"}\n", msg)
	return nil
}

type ValidateFn func(cfg *Config) error

func DefaultValidateOpenAIFunc(cfg *Config) error {
	url := "https://api.openai.com/v1/models"
	return DoValidate(cfg.APIKey, url, "/tools/openai-model-provider/validate", "Invalid OpenAI Credentials")
}

func DoValidate(apiKey, urlStr, loggerPath, invalidCredsMsg string) error {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}
	defer resp.Body.Close()

	var modelsResp struct {
		Object string `json:"object"`
		Data   []any  `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	if len(modelsResp.Data) == 0 {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	slog.Info("Credentials are valid", "logger", loggerPath)
	return nil
}
