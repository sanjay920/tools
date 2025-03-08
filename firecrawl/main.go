package main

import (
	"context"
	"fmt"
	"os"

	"github.com/obot-platform/tools/firecrawl/cmd"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	var (
		result string
		err    error
		ctx    = context.Background()
	)

	switch command {
	case "scrapeUrl":
		result, err = cmd.Scrape(ctx,
			os.Getenv("FIRECRAWL_API_KEY"),
			os.Getenv("FIRECRAWL_URL_TO_SCRAPE"),
		)
	default:
		err = fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(result)
}

