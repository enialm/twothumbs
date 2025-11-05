// File: internal/templates/digests/quarterly.go

// This file contains the Block Kit template for quarterly digests.

package digests

import (
	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func BuildQuarterlyDigestBlocks(digests []models.QuarterlyDigestData, quarterLabel string) []map[string]any {
	var blocks []map[string]any

	// Main header
	blocks = append(blocks,
		utils.Spacer(),
		map[string]any{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Quarterly Digest  📊",
			},
		},
		map[string]any{
			"type": "context",
			"elements": []map[string]any{
				{
					"type": "plain_text",
					"text": quarterLabel,
				},
			},
		},
		utils.Spacer(),
	)

	for _, d := range digests {
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
		if d.GraphURL != "" {
			blocks = append(blocks,
				map[string]any{
					"type":      "image",
					"image_url": d.GraphURL,
					"alt_text":  d.Origin,
				},
			)
		}
		blocks = append(blocks,
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
