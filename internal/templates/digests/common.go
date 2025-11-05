// File: internal/templates/digests/common.go

// This file contains common templates for digest blocks.

package digests

func DigestFooter() []map[string]any {
	return []map[string]any{
		{
			"type": "divider",
		},
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type": "button",
					"text": map[string]any{
						"type": "plain_text",
						"text": "Explore Feedback",
					},
					"action_id": "explore_feedback",
				},
				{
					"type": "button",
					"text": map[string]any{
						"type": "plain_text",
						"text": "Report an Issue",
					},
					"action_id": "report_issue",
				},
			},
		},
	}
}
