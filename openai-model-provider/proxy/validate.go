package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func init() {
	log.SetFlags(0)
}

func logInfo(loggerPath string, msg string) {
	log.Printf("time=%q level=info msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
}

func logError(loggerPath string, msg string) {
	log.Printf("time=%q level=error msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
}

func handleValidationError(loggerPath, msg string) error {
	logError(loggerPath, msg)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	if resp.StatusCode != 200 {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	var modelsResp struct {
		Object string `json:"object"`
		Data   []any  `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}
	if len(modelsResp.Data) == 0 {
		return handleValidationError(loggerPath, invalidCredsMsg)
	}

	logInfo(loggerPath, "Credentials are valid")
	fmt.Println("{\"message\": \"Credentials are valid\"}")
	return nil
}
