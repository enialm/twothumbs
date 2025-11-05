// File: internal/templates/modals/report.go

// This file contains the logic for the Slack App Report modal.

package modals

import "fmt"

func ReportModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Report an Issue  🛠️",
		},
		"submit": map[string]any{
			"type": "plain_text",
			"text": "Submit",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": []map[string]any{
			{
				"type": "input",
				"element": map[string]any{
					"type":      "plain_text_input",
					"action_id": "message-text",
					"multiline": true,
					"placeholder": map[string]any{
						"type": "plain_text",
						"text": "Pour your heart out",
					},
				},
				"label": map[string]any{
					"type": "plain_text",
					"text": "Issue Description",
				},
			},
			{
				"type": "input",
				"element": map[string]any{
					"type":      "email_text_input",
					"action_id": "email-address",
					"placeholder": map[string]any{
						"type": "plain_text",
						"text": "Let us take care of you",
					},
				},
				"label": map[string]any{
					"type": "plain_text",
					"text": "Email Address",
				},
			},
		},
	}
}

func ReportSentModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "We got you  🫡",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": "Your issue report has been submitted. We shall read it, decide what to do, and contact you afterwards. Hang on tight!",
				},
			},
		},
	}
}

func ReportErrorModal(supportEmail string) map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Uh-oh  🫣",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": fmt.Sprintf("Something went horribly wrong. We beg your pardon, and know this sucks: please retype everything and send it directly to %s.", supportEmail),
				},
			},
		},
	}
}
