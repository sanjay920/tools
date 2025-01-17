package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/obot-platform/tools/openai-model-provider/api"
)

func handleValidationError(loggerPath, msg string) error {
	slog.Error(msg, "logger", loggerPath)
	fmt.Printf("{\"error\": \"%s\"}\n", msg)
	return nil
}

func DoValidate(cfg *Config, loggerPath, invalidCredsMsg string) error {
	scheme := "https"
	if !cfg.UseTLS {
		scheme = "http"
	}

	url := fmt.Sprintf("%s://%s%s/v1/models", scheme, cfg.UpstreamHost, cfg.PathPrefix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	var modelsResp api.ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	if modelsResp.Object != "list" || len(modelsResp.Data) == 0 {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	return nil
}

func DefaultValidateOpenAIFunc(cfg *Config) error {
	return DoValidate(cfg, "/tools/openai-model-provider/validate", "Invalid OpenAI Credentials")
}
