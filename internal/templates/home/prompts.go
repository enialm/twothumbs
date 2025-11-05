// File: internal/templates/home/prompts.go

// This file contains the logic for the Slack App Home Prompts tab.

package home

import (
	"fmt"

	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func PromptsBlocks(prompts []models.Prompt, maxPromptCount int) []map[string]any {
	blocks := []map[string]any{
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Explore Feedback"},
					"action_id": "home-explore",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Prompts"},
					"action_id": "home-prompts",
					"style":     "primary",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Settings"},
					"action_id": "home-settings",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Support"},
					"action_id": "home-support",
				},
			},
		},
		utils.Spacer(),
		{
			"type": "divider",
		},
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Prompts  🙌",
			},
		},
		utils.Spacer(),
		{
			"type": "context",
			"elements": []map[string]any{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("_%d/%d prompts in use_", len(prompts), maxPromptCount),
				},
			},
		},
		utils.Spacer(),
		{
			"type": "divider",
		},
	}

	if len(prompts) == 0 {
		blocks = append(blocks,
			utils.Spacer(),
			map[string]any{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": "_No prompts to display_",
				},
			},
			utils.Spacer(),
		)
	} else {
		for i, p := range prompts {
			blocks = append(blocks,
				map[string]any{
					"type": "section",
					"text": map[string]any{
						"type": "mrkdwn",
						"text": fmt.Sprintf(
							"Origin: %s\nCategory: %s\n\n>_%s_",
							p.Origin, p.Category, p.Prompt,
						),
					},
				},
				map[string]any{
					"type": "actions",
					"elements": []map[string]any{
						{
							"type":      "button",
							"text":      map[string]any{"type": "plain_text", "text": "Delete"},
							"action_id": "modal-delete-prompt",
							"value":     fmt.Sprintf("%d", p.ID),
						},
					},
				},
			)

			if i < len(prompts)-1 {
				blocks = append(blocks, map[string]any{
					"type": "divider",
				})
			}
		}
	}

	return blocks
}
