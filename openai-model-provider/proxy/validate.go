package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ValidateFn func(cfg *Config) error

func DefaultValidateOpenAIFunc(cfg *Config) error {
	url := "https://api.openai.com/v1/models"
	return doValidate(cfg.APIKey, url)
}

func ValidateDeepSeekAPIKey(cfg *Config) error {
	url := "https://api.deepseek.com/v1/models"
	return doValidate(cfg.APIKey, url)
}

func doValidate(apiKey, urlStr string) error {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to %q: %w", urlStr, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return fmt.Errorf("authentication failed: %s (type=%s)", errResp.Error.Message, errResp.Error.Type)
		}
		return fmt.Errorf("API validation failed; status=%d body=%q", resp.StatusCode, string(body))
	}

	var modelsResp struct {
		Object string `json:"object"`
		Data   []any  `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return fmt.Errorf("failed to parse model list: %w. Raw body=%q", err, string(body))
	}
	if len(modelsResp.Data) == 0 {
		return fmt.Errorf("no models found in response")
	}
	return nil
}
