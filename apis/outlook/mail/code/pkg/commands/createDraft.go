package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/util"
)

func CreateDraft(ctx context.Context, info graph.DraftInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	draft, err := graph.CreateDraft(ctx, c, info)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	fmt.Printf("Draft created successfully. Draft ID: %s\n", util.Deref(draft.GetId()))
	return nil
}