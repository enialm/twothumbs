// File: internal/templates/digests/monthly.go

// This file contains the Block Kit template for monthly digests.

package digests

import (
	"fmt"

	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func BuildMonthlyDigestBlocks(digests []models.MonthlyDigestData, monthLabel string) []map[string]any {
	var blocks []map[string]any

	// Main header
	blocks = append(blocks,
		utils.Spacer(),
		map[string]any{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Monthly Digest  🌙",
			},
		},
		map[string]any{
			"type": "context",
			"elements": []map[string]any{
				{
					"type": "plain_text",
					"text": monthLabel,
				},
			},
		},
		utils.Spacer(),
	)

	var lastOrigin string
	for _, d := range digests {
		// Origin header and graph
		if d.Origin != lastOrigin {
			blocks = append(blocks,
				map[string]any{
					"type": "divider",
				},
				map[string]any{
					"type": "header",
					"text": map[string]any{
						"type": "plain_text",
						"text": d.Origin,
					},
				},
			)
			lastOrigin = d.Origin
		}

		// Category section
		blocks = append(blocks,
			utils.Spacer(),
			map[string]any{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": MonthlyBlockText(
						d.Category,
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
		)
		if d.GraphURL != "" {
			blocks = append(blocks,
				map[string]any{
					"type":      "image",
					"image_url": d.GraphURL,
					"alt_text":  d.Category,
				},
			)
		}
		blocks = append(blocks,
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

func MonthlyBlockText(
	category string,
	score int,
	scoreDelta string,
	nResponses int,
	nResponsesDelta string,
	nComments int,
	nCommentsDelta string,
) string {
	return fmt.Sprintf(
		"*%s*\n\n\n👍   %d%% (%s%%)    👋   %d (%s%%)    💬   %d (%s%%)",
		category,
		score,
		scoreDelta,
		nResponses,
		nResponsesDelta,
		nComments,
		nCommentsDelta,
	)
}
