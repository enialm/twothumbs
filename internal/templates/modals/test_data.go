// File: internal/templates/modals/test_data.go

// This file contains the test data modal template.

package modals

import (
	"fmt"

	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func TestDataModal(data []models.RawData) map[string]any {
	var blocks []map[string]any

	if len(data) == 0 {
		blocks = append(blocks,
			utils.Spacer(),
			map[string]any{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "No data to display",
				},
			},
			utils.Spacer(),
		)
	} else {
		blocks = append(blocks,
			map[string]any{
				"type": "context",
				"elements": []map[string]any{
					{
						"type": "mrkdwn",
						"text": "_Kindly note: Only the ten most recent events are shown._",
					},
				},
			},
			utils.Spacer(),
		)

		for i, vals := range data {
			comment := " "
			if vals.Comment != nil {
				comment = *vals.Comment
			}
			blocks = append(blocks, map[string]any{
				"type": "section",
				"fields": []map[string]any{
					{"type": "plain_text", "text": "origin:"},
					{"type": "plain_text", "text": vals.Origin},
					{"type": "plain_text", "text": "category:"},
					{"type": "plain_text", "text": vals.Category},
					{"type": "plain_text", "text": "prompt:"},
					{"type": "plain_text", "text": vals.Prompt},
				},
			})
			blocks = append(blocks, map[string]any{
				"type": "section",
				"fields": []map[string]any{
					{"type": "plain_text", "text": "thumb_up:"},
					{"type": "plain_text", "text": fmt.Sprintf("%v", vals.ThumbUp)},
					{"type": "plain_text", "text": "comment:"},
					{"type": "plain_text", "text": comment},
					{"type": "plain_text", "text": "user_id:"},
					{"type": "plain_text", "text": vals.UserID},
				},
			})
			blocks = append(blocks, utils.Spacer())
			blocks = append(blocks, map[string]any{
				"type": "context",
				"elements": []map[string]any{
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("Sent %s ago  (%s UTC)", utils.TimeToAgo(vals.CreatedAt), vals.CreatedAt.Format("2006-01-02 15:04")),
					},
				},
			})
			if i < len(data)-1 {
				blocks = append(blocks,
					utils.Spacer(),
					map[string]any{
						"type": "divider",
					},
				)
			}
		}
	}

	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Test Data",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": blocks,
	}
}
