package commands

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"

    "github.com/gptscript-ai/tools/elevenlabs/pkg/client"
)

func TextToSpeech(ctx context.Context, text, voiceID, outputFile string) error {
    url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
    
    payload := strings.NewReader(fmt.Sprintf(`{
        "text": "%s",
        "model_id": "eleven_monolingual_v1",
        "voice_settings": {
            "stability": 0.5,
            "similarity_boost": 0.5
        }
    }`, text))

    req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Add("Accept", "audio/mpeg")
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("xi-api-key", client.GetAPIKey())

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to make request: %w", err)
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(res.Body)
        return fmt.Errorf("request failed with status %d: %s", res.StatusCode, string(body))
    }

    out, err := os.Create(outputFile)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer out.Close()

    _, err = io.Copy(out, res.Body)
    if err != nil {
        return fmt.Errorf("failed to write response to file: %w", err)
    }

    fmt.Printf("Successfully generated speech and saved to %s\n", outputFile)
    return nil
}

