// File: internal/templates/digests/weekly.go

// This file contains the Block Kit template for weekly digests.

package digests

import (
	"fmt"

	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func BuildWeeklyDigestBlocks(digests []models.WeeklyDigestData) []map[string]any {
	var blocks []map[string]any

	// Main header
	blocks = append(blocks,
		utils.Spacer(),
		map[string]any{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Weekly Digest  🚀",
			},
		},
		utils.Spacer(),
	)

	var lastOrigin, lastCategory string
	for _, d := range digests {
		// Origin header
		if d.Origin != lastOrigin {
			blocks = append(blocks,
				map[string]any{
					"type": "header",
					"text": map[string]any{
						"type": "plain_text",
						"text": d.Origin,
					},
				},
				utils.Spacer(),
			)
			lastOrigin = d.Origin
			lastCategory = ""
		}

		// Category header
		if d.Category != lastCategory {
			blocks = append(blocks,
				map[string]any{
					"type": "divider",
				},
				map[string]any{
					"type": "section",
					"text": map[string]any{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*%s*", d.Category),
					},
				},
				utils.Spacer(),
			)
			lastCategory = d.Category
		}

		// Prompt section
		blocks = append(blocks,
			map[string]any{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": utils.FormatDigestStats(
						d.Prompt,
						d.Score,
						d.ScoreDelta,
						d.NResponses,
						d.NResponsesDelta,
						d.NComments,
						d.NCommentsDelta,
					),
				},
			},
			utils.Spacer(),
			map[string]any{
				"type": "rich_text",
				"elements": []map[string]any{
					{
						"type": "rich_text_quote",
						"elements": []map[string]any{
							{
								"type": "text",
								"text": d.Digest,
							},
						},
					},
				},
			},
			utils.Spacer(),
		)
	}

	// Footer
	blocks = append(
		blocks,
		DigestFooter()...,
	)

	return blocks
}
