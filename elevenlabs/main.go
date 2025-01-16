package main

import (
    "context"
    "fmt"
    "os"

    "github.com/gptscript-ai/tools/elevenlabs/pkg/commands"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Command required")
        os.Exit(1)
    }

    var err error
    ctx := context.Background()

    switch os.Args[1] {
    case "text-to-speech":
        if len(os.Args) < 4 {
            fmt.Println("Usage: text-to-speech <text> <voice_id> <output_file>")
            os.Exit(1)
        }
        err = commands.TextToSpeech(ctx, os.Args[2], os.Args[3], os.Args[4])
    default:
        fmt.Printf("Unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
