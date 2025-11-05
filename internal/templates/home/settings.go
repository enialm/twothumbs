// File: internal/templates/home/tabs/settings.go

// This file contains the logic for the Slack App Home Settings tab.

package home

import (
	"fmt"
	"twothumbs/internal/utils"
)

func SettingsBlocks(channel, apiKey string) []map[string]any {
	channelSelect := map[string]any{
		"type": "channels_select",
		"placeholder": map[string]any{
			"type": "plain_text",
			"text": "Select a channel",
		},
		"action_id": "set-channel",
	}
	if channel != "" {
		channelSelect["initial_channel"] = channel
	}

	channelActions := []map[string]any{channelSelect}
	if channel != "" {
		channelActions = append(channelActions, map[string]any{
			"type":      "button",
			"text":      map[string]any{"type": "plain_text", "text": "Clear"},
			"action_id": "clear-channel",
			"value":     channel,
			"confirm": map[string]any{
				"title": map[string]any{
					"type": "plain_text",
					"text": "Clear Output Channel?",
				},
				"text": map[string]any{
					"type": "plain_text",
					"text": "No digests shall be delivered without an output channel. Are you certain you want to clear the output channel?",
				},
				"confirm": map[string]any{
					"type": "plain_text",
					"text": "Clear",
				},
				"deny": map[string]any{
					"type": "plain_text",
					"text": "Cancel",
				},
			},
		})
	}

	return []map[string]any{
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
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Settings"},
					"action_id": "home-settings",
					"style":     "primary",
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
				"text": "Output Channel  📢",
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Select the channel where digests are sent:",
			},
		},
		{
			"type":     "actions",
			"elements": channelActions,
		},
		{
			"type": "context",
			"elements": []map[string]any{
				{
					"type": "mrkdwn",
					"text": "_Please note the Two Thumbs app must be invited to the channel (use Slack's /invite command)._",
				},
			},
		},
		utils.Spacer(),
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "API Key  🔑",
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "plain_text",
				"text": fmt.Sprintf("Your current API key: %s", apiKey),
			},
		},
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Rotate API Key"},
					"action_id": "rotate-api-key",
					"confirm": map[string]any{
						"title": map[string]any{
							"type": "plain_text",
							"text": "Rotate API Key?",
						},
						"text": map[string]any{
							"type": "plain_text",
							"text": "Are you sure you want to rotate your API key? This will invalidate your current key and generate a new one.",
						},
						"confirm": map[string]any{
							"type": "plain_text",
							"text": "Rotate",
						},
						"deny": map[string]any{
							"type": "plain_text",
							"text": "Cancel",
						},
					},
				},
			},
		},
		utils.Spacer(),
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Test Data  ⚗️",
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "plain_text",
				"text": "View feedback data with in_production = false:",
			},
		},
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "View Test Data"},
					"action_id": "view-test-data",
				},
			},
		},
	}
}
