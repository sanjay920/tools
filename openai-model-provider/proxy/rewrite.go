package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	openai "github.com/gptscript-ai/chat-completion-client"
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

	var models openai.ModelsList
	if err := json.NewDecoder(body).Decode(&models); err != nil {
		return fmt.Errorf("failed to decode models response: %w", err)
	}

	for i, model := range models.Models {
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
		models.Models[i] = model
	}

	b, err := json.Marshal(models)
	if err != nil {
		return fmt.Errorf("failed to marshal models response: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return nil
}

func RewriteAllModelsWithUsage(usage string) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return nil
		}
		originalBody := resp.Body
		defer originalBody.Close()

		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(originalBody)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gzReader.Close()
			resp.Header.Del("Content-Encoding")
			originalBody = gzReader
		}

		var models openai.ModelsList
		if err := json.NewDecoder(originalBody).Decode(&models); err != nil {
			return fmt.Errorf("failed to decode models response: %w", err)
		}

		for i, model := range models.Models {
			if model.Metadata == nil {
				model.Metadata = make(map[string]string)
			}
			model.Metadata["usage"] = usage
			models.Models[i] = model
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
