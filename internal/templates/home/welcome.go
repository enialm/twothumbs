// File: internal/templates/home/welcome.go

// This file contains the logic for the welcome view of the Slack App Home.

package home

import "twothumbs/internal/utils"

func WelcomeBlocks() []map[string]any {
	return []map[string]any{
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Welcome  🎉",
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "plain_text",
				"text": "To get started, link your Slack workspace to your Two Thumbs account. Just submit the activation code found in the welcome email and we are good to go.",
			},
		},
		{
			"dispatch_action": true,
			"type":            "input",
			"element": map[string]any{
				"type":        "plain_text_input",
				"action_id":   "link-workspace",
				"placeholder": map[string]any{"type": "plain_text", "text": "Activation code"},
				"max_length":  14, // TWTH-XXXX-YYYY
			},
			"label": map[string]any{
				"type": "plain_text",
				"text": " ",
			},
		},
	}
}
