// File: internal/templates/modals/add_new_prompt.go

// This file contains the modal template for adding a new prompt.

package modals

import (
	"fmt"

	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func DeletePromptModal(p models.Prompt) map[string]any {
	return map[string]any{
		"type":             "modal",
		"callback_id":      "delete-prompt",
		"private_metadata": fmt.Sprintf("%d", p.ID), // To identify the prompt
		"title": map[string]any{
			"type": "plain_text",
			"text": "Delete Prompt?  🛑",
		},
		"submit": map[string]any{
			"type": "plain_text",
			"text": "Delete",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "You are about the delete the following prompt and ALL RELATED FEEDBACK DATA. This action cannot be undone. Are you absolutely sure you wish to proceed?",
				},
			},
			utils.Spacer(),
			{
				"type": "divider",
			},
			utils.Spacer(),
			{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": fmt.Sprintf(
						"Origin: %s\nCategory: %s\n\n>_%s_",
						p.Origin, p.Category, p.Prompt,
					),
				},
			},
			utils.Spacer(),
			{
				"type": "divider",
			},
			{
				"type": "input",
				"element": map[string]any{
					"type":        "plain_text_input",
					"action_id":   "delete-confirm",
					"placeholder": map[string]any{"type": "plain_text", "text": "YES"},
				},
				"label": map[string]any{
					"type": "plain_text",
					"text": "Type 'YES' to proceed",
				},
			},
			utils.Spacer(),
		},
	}
}

func DeletePromptSuccessModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Done  ✅",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "The prompt and all related feedback were deleted successfully.",
				},
			},
		},
	}
}

func DeletePromptErrorModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Ouch  🤕",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "There was an error deleting prompt or feedback data. Please try again and contact our support if the error persists. We are truly sorry for the inconvenience.",
				},
			},
		},
	}
}

func DeletePromptNoYesModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Confirmation Required  🤖",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "You must type 'YES' to confirm deletion. The prompt and feedback data were NOT deleted.",
				},
			},
		},
	}
}
