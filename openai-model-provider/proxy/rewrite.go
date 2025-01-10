package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func DefaultRewriteModelsResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	defer resp.Body.Close()

	var body io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		resp.Header.Del("Content-Encoding")
		body = gzReader
	}

	var models struct {
		Object string `json:"object"`
		Data   []struct {
			ID       string            `json:"id"`
			Object   string            `json:"object"`
			OwnedBy  string            `json:"owned_by"`
			Metadata map[string]string `json:"metadata,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(body).Decode(&models); err != nil {
		return fmt.Errorf("failed to decode models response: %w", err)
	}

	for i, model := range models.Data {
		if model.Metadata == nil {
			model.Metadata = make(map[string]string)
		}
		switch {
		case strings.HasPrefix(model.ID, "gpt-"),
			strings.HasPrefix(model.ID, "ft:gpt-"),
			strings.HasPrefix(model.ID, "o1-"),
			strings.HasPrefix(model.ID, "ft:o1-"):
			model.Metadata["usage"] = "llm"
		case strings.HasPrefix(model.ID, "text-embedding-"),
			strings.HasPrefix(model.ID, "ft:text-embedding-"):
			model.Metadata["usage"] = "text-embedding"
		case strings.HasPrefix(model.ID, "dall-e"),
			strings.HasPrefix(model.ID, "ft:dall-e"):
			model.Metadata["usage"] = "image-generation"
		}
		models.Data[i] = model
	}

	b, err := json.Marshal(models)
	if err != nil {
		return fmt.Errorf("failed to marshal models response: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return nil
}

func RewriteAllModelsWithUsage(usage string, filter ...func(string) bool) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return nil
		}

		defer resp.Body.Close()

		var body io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gzReader.Close()
			resp.Header.Del("Content-Encoding")
			body = gzReader
		}

		var models struct {
			Object string `json:"object"`
			Data   []struct {
				ID       string            `json:"id"`
				Object   string            `json:"object"`
				OwnedBy  string            `json:"owned_by"`
				Metadata map[string]string `json:"metadata,omitempty"`
			} `json:"data"`
		}

		if err := json.NewDecoder(body).Decode(&models); err != nil {
			return fmt.Errorf("failed to decode models response: %w", err)
		}

		for i, model := range models.Data {
			shouldMark := true
			if len(filter) > 0 && filter[0] != nil {
				shouldMark = filter[0](model.ID)
			}

			if shouldMark {
				if model.Metadata == nil {
					model.Metadata = make(map[string]string)
				}
				model.Metadata["usage"] = usage
				models.Data[i] = model
			}
		}

		b, err := json.Marshal(models)
		if err != nil {
			return fmt.Errorf("failed to marshal models response: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(b))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		return nil
	}
}
