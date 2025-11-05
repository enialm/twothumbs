// File: internal/templates/modals/explore.go

// This file contains the logic for the Slack App Explore modal.

package modals

import (
	"fmt"

	"twothumbs/internal/config"
	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

func ExploreModal() map[string]any {
	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Explore Feedback",
		},
		"blocks": []map[string]any{
			{
				"type": "header",
				"text": map[string]any{
					"type": "plain_text",
					"text": "Quick Actions  ⚡️",
				},
			},
			{
				"type": "actions",
				"elements": []map[string]any{
					{
						"type":      "button",
						"text":      map[string]any{"type": "plain_text", "text": "Latest Comments"},
						"action_id": "modal-comments",
					},
					{
						"type":      "button",
						"text":      map[string]any{"type": "plain_text", "text": "Top Issues"},
						"action_id": "modal-top-issues",
					},
					{
						"type":      "button",
						"text":      map[string]any{"type": "plain_text", "text": "7-Day Stats"},
						"action_id": "modal-stats-7d",
					},
					{
						"type":      "button",
						"text":      map[string]any{"type": "plain_text", "text": "30-Day Stats"},
						"action_id": "modal-stats-30d",
					},
				},
			},
			utils.Spacer(),
			{
				"type": "header",
				"text": map[string]any{
					"type": "plain_text",
					"text": "Slow Actions  🐢",
				},
			},
			{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": "Navigate to the Two Thumbs home view to query feedback using different filters. You should find Two Thumbs in the Apps section of the sidebar.",
				},
			},
			utils.Spacer(),
		},
	}
}

func ExploreCommentsModal(title string, cfg *config.InteractConfig, results []models.ExploreCommentsResult) map[string]any {
	blocks := []map[string]any{}

	if len(results) == 0 {
		blocks = append(blocks,
			utils.Spacer(),
			map[string]any{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "No comments to display",
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
						"text": fmt.Sprintf("_Kindly note: Only the %d most recent comments are shown._", cfg.NComments),
					},
				},
			},
			utils.Spacer(),
		)
		for i, r := range results {
			blocks = append(blocks,
				map[string]any{
					"type": "section",
					"text": map[string]any{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*%s / %s*\n\n_%s_\n\n> %s", r.Origin, r.Category, r.Prompt, r.Comment),
					},
				},
				map[string]any{
					"type": "context",
					"elements": []map[string]any{
						{
							"type": "plain_text",
							"text": fmt.Sprintf("%s ago", utils.TimeToAgo(r.Timestamp)),
						},
					},
				},
				utils.Spacer(),
			)
			if i < len(results)-1 {
				blocks = append(blocks, map[string]any{"type": "divider"})
			}
		}
	}

	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": title,
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": blocks,
	}
}

func TopIssuesModal(issues []models.IssueReport) map[string]any {
	blocks := []map[string]any{}

	if len(issues) == 0 {
		blocks = []map[string]any{
			utils.Spacer(),
			{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "No issues to display",
				},
			},
			utils.Spacer(),
		}
	} else {
		blocks = append(blocks,
			map[string]any{
				"type": "context",
				"elements": []map[string]any{
					{
						"type": "mrkdwn",
						"text": "_Please note: The issues are based on feedback gathered during the last 30 days._",
					},
				},
			},
			utils.Spacer(),
		)
		for i, issue := range issues {
			blocks = append(blocks,
				map[string]any{
					"type": "header",
					"text": map[string]any{
						"type": "plain_text",
						"text": issue.Origin,
					},
				},
				utils.Spacer(),
				map[string]any{
					"type": "section",
					"text": map[string]any{
						"type": "mrkdwn",
						"text": utils.FormatIssues(issue.Report),
					},
				},
				utils.Spacer(),
			)
			if i < len(issues)-1 {
				blocks = append(blocks, map[string]any{"type": "divider"})
			}
		}
	}

	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": "Top Issues ❗️",
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": blocks,
	}
}

func StatsModal(title string, stats []models.StatsData) map[string]any {
	blocks := []map[string]any{}

	if len(stats) == 0 {
		blocks = []map[string]any{
			utils.Spacer(),
			{
				"type": "section",
				"text": map[string]any{
					"type": "plain_text",
					"text": "No stats to display",
				},
			},
			utils.Spacer(),
		}
	} else {
		var lastOrigin, lastCategory string
		for _, s := range stats {
			// Add a header for each origin
			if s.Origin != lastOrigin {
				blocks = append(blocks,
					map[string]any{
						"type": "header",
						"text": map[string]any{
							"type": "plain_text",
							"text": s.Origin,
						},
					},
					utils.Spacer(),
					map[string]any{
						"type": "divider",
					},
				)
				lastOrigin = s.Origin
				lastCategory = ""
			}

			// Add a section for each prompt
			if s.Category != lastCategory {
				blocks = append(blocks,
					map[string]any{
						"type": "section",
						"text": map[string]any{
							"type": "mrkdwn",
							"text": utils.FormatModalStats(
								s.Category,
								s.Prompt,
								s.Score,
								s.ScoreDelta,
								s.NResponses,
								s.NResponsesDelta,
								s.NComments,
								s.NCommentsDelta,
								true, // printCategory
							),
						},
					},
				)
				lastCategory = s.Category
			} else {
				blocks = append(blocks,
					map[string]any{
						"type": "section",
						"text": map[string]any{
							"type": "mrkdwn",
							"text": utils.FormatModalStats(
								s.Category,
								s.Prompt,
								s.Score,
								s.ScoreDelta,
								s.NResponses,
								s.NResponsesDelta,
								s.NComments,
								s.NCommentsDelta,
								false, // printCategory
							),
						},
					},
				)
			}
			blocks = append(blocks, utils.Spacer())
		}
	}

	return map[string]any{
		"type": "modal",
		"title": map[string]any{
			"type": "plain_text",
			"text": title,
		},
		"close": map[string]any{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": blocks,
	}
}
