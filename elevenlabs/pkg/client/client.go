package client

import (
    "os"
)

const CredentialEnv = "ELEVENLABS_API_KEY"

func GetAPIKey() string {
    return os.Getenv(CredentialEnv)
}

