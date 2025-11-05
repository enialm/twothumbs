// File: internal/templates/modals/report.go

// This file contains the logic for the Slack App Support modal.

package modals

import "fmt"

func SupportModal() map[string]any {
	return map[string]any{
		"type":        "modal",
		"callback_id": "contact-support",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Contact Support  🦸",
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
						"text": "Write anything",
					},
				},
				"label": map[string]any{
					"type": "plain_text",
					"text": "How can we help you?",
				},
			},
			{
				"type": "input",
				"element": map[string]any{
					"type":      "email_text_input",
					"action_id": "email-address",
					"placeholder": map[string]any{
						"type": "plain_text",
						"text": "Let us take care of everything",
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

func ContactSupportSentModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "We're on it  🤓",
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
					"text": "Your message has been submitted. We shall read it and contact you soon. In the meantime, have a lovely day!",
				},
			},
		},
	}
}

func ContactSupportErrorModal(supportEmail string) map[string]any {
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
