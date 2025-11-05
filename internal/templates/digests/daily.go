// File: internal/templates/digests/daily.go

// This file contains the Block Kit template for daily digests.

package digests

import (
	"fmt"
	"twothumbs/internal/models"
)

func BuildDailyDigestBlocks(digests []models.DailyDigestData) []map[string]any {
	var blocks []map[string]any

	// Main header
	blocks = append(blocks,
		map[string]any{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Daily Digest  ☀️",
			},
		},
	)

	// Content
	if len(digests) == 0 {
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "_No new comments in the last 24 hours_",
			},
		})
	} else {
		for _, d := range digests {
			commentWord := "comments"
			if d.NComments == 1 {
				commentWord = "comment"
			}
			blocks = append(blocks,
				map[string]any{
					"type": "section",
					"text": map[string]any{
						"type": "mrkdwn",
						"text": DailyBlockText(d.Origin, d.NComments, commentWord),
					},
				},
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
		}
	}

	// Footer
	blocks = append(
		blocks,
		DigestFooter()...,
	)

	return blocks
}

func DailyBlockText(
	origin string,
	nComments int,
	commentWord string,
) string {
	return fmt.Sprintf(
		"*%s* _(%d new %s)_",
		origin,
		nComments,
		commentWord,
	)
}
