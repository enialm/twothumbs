// File: internal/templates/home/tabs/explore.go

// This file contains the logic for the Slack App Home Explore tab.

package home

import (
	"fmt"
	"time"

	"twothumbs/internal/utils"
)

func ExploreBlocks(
	origins []string,
	categories []string,
	prompts []string,
	dateFrom time.Time,
	dateTo time.Time,
	dateInitial bool,
	numFeedbackItems int,
	numComments int,
	selectedOrigin string,
	selectedCategory string,
	selectedPrompt string,
	selectedThumb string,
	expanded bool,
) []map[string]any {
	if expanded {
		return exploreBlocksExpanded(
			origins, categories, prompts, dateFrom, dateTo, dateInitial,
			numFeedbackItems, numComments, selectedOrigin, selectedCategory, selectedPrompt, selectedThumb,
		)
	}
	return exploreBlocksDefault()
}

func exploreBlocksDefault() []map[string]any {
	blocks := []map[string]any{
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Explore Feedback"},
					"action_id": "home-explore",
					"style":     "primary",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Prompts"},
					"action_id": "home-prompts",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Settings"},
					"action_id": "home-settings",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Support"},
					"action_id": "home-support",
				},
			},
		},
		utils.Spacer(),
		{
			"type": "divider",
		},
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Quick Actions  ⚡️",
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Read the latest comments, see AI-identified issues, or quickly review stats. Cheers!",
			},
		},
		utils.Spacer(),
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Latest Comments"},
					"action_id": "home-comments",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Top Issues"},
					"action_id": "home-top-issues",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "7-Day Stats"},
					"action_id": "home-stats-7d",
				},
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "30-Day Stats"},
					"action_id": "home-stats-30d",
				},
			},
		},
		utils.Spacer(),
		{
			"type": "divider",
		},
		utils.Spacer(),
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Stats, Comments, Raw Data  🕵",
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "plain_text",
				"text": "Dive deep into feedback by viewing stats, reading comments, or exporting feedback as a CSV file for additional analysis. Use the provided filters to narrow down the data as needed. Have fun!",
			},
		},
		utils.Spacer(),
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "Begin"},
					"action_id": "home-expand",
				},
			},
		},
	}
	return blocks
}

func exploreBlocksExpanded(
	origins []string,
	categories []string,
	prompts []string,
	dateFrom time.Time,
	dateTo time.Time,
	dateInitial bool,
	numFeedbackItems int,
	numComments int,
	selectedOrigin string,
	selectedCategory string,
	selectedPrompt string,
	selectedThumb string,
) []map[string]any {
	blocks := []map[string]any{
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":      "button",
					"text":      map[string]any{"type": "plain_text", "text": "< Back"},
					"action_id": "home-explore",
				},
			},
		},
		utils.Spacer(),
		{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "*Filters*",
			},
		},
	}

	// Origin filter row
	originRow := []map[string]any{
		staticSelectBlock(
			"select-origin",
			fmt.Sprintf("Origin (%d)", len(origins)),
			origins,
			selectedOrigin,
		),
	}
	if selectedOrigin != "" {
		originRow = append(originRow, map[string]any{
			"type":      "button",
			"text":      map[string]any{"type": "plain_text", "text": "Clear"},
			"action_id": "clear-origin",
			"style":     "danger",
		})
	}
	blocks = append(blocks, map[string]any{
		"type":     "actions",
		"elements": originRow,
	})

	// Category filter row
	categoryRow := []map[string]any{
		staticSelectBlock(
			"select-category",
			fmt.Sprintf("Category (%d)", len(categories)),
			categories,
			selectedCategory,
		),
	}
	if selectedCategory != "" {
		categoryRow = append(categoryRow, map[string]any{
			"type":      "button",
			"text":      map[string]any{"type": "plain_text", "text": "Clear"},
			"action_id": "clear-category",
			"style":     "danger",
		})
	}
	blocks = append(blocks, map[string]any{
		"type":     "actions",
		"elements": categoryRow,
	})

	// Prompt filter row
	promptRow := []map[string]any{
		staticSelectBlock(
			"select-prompt",
			fmt.Sprintf("Prompt (%d)", len(prompts)),
			prompts,
			selectedPrompt,
		),
	}
	if selectedPrompt != "" {
		promptRow = append(promptRow, map[string]any{
			"type":      "button",
			"text":      map[string]any{"type": "plain_text", "text": "Clear"},
			"action_id": "clear-prompt",
			"style":     "danger",
		})
	}
	blocks = append(blocks, map[string]any{
		"type":     "actions",
		"elements": promptRow,
	})

	// Thumb filter row
	thumbOptions := []string{"Up", "Down"}
	thumbRow := []map[string]any{
		staticSelectBlock(
			"select-thumb",
			"Thumb",
			thumbOptions,
			selectedThumb,
		),
	}
	if selectedThumb != "" {
		thumbRow = append(thumbRow, map[string]any{
			"type":      "button",
			"text":      map[string]any{"type": "plain_text", "text": "Clear"},
			"action_id": "clear-thumb",
			"style":     "danger",
		})
	}
	blocks = append(blocks, map[string]any{
		"type":     "actions",
		"elements": thumbRow,
	})

	// Date range section
	blocks = append(blocks, map[string]any{
		"type": "section",
		"text": map[string]any{"type": "mrkdwn", "text": "*Date Range (UTC)*"},
	})

	dateRow := []map[string]any{
		{
			"type":         "datepicker",
			"initial_date": dateFrom.Format("2006-01-02"),
			"placeholder": map[string]any{
				"type": "plain_text",
				"text": "Start date",
			},
			"action_id": "stats-date-from",
		},
		{
			"type":         "datepicker",
			"initial_date": dateTo.Format("2006-01-02"),
			"placeholder": map[string]any{
				"type": "plain_text",
				"text": "End date",
			},
			"action_id": "stats-date-to",
		},
	}
	if !dateInitial {
		dateRow = append(dateRow, map[string]any{
			"type":      "button",
			"text":      map[string]any{"type": "plain_text", "text": "Reset"},
			"action_id": "clear-dates",
			"style":     "danger",
		})
	}
	blocks = append(blocks, map[string]any{
		"type":     "actions",
		"elements": dateRow,
	})

	blocks = append(blocks, []map[string]any{
		utils.Spacer(),
		{
			"type": "context",
			"elements": []map[string]any{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf(
						"_%d matching feedback %s with %d %s_",
						numFeedbackItems,
						func() string {
							if numFeedbackItems == 1 {
								return "item"
							}
							return "items"
						}(),
						numComments,
						func() string {
							if numComments == 1 {
								return "comment"
							}
							return "comments"
						}(),
					),
				},
			},
		},
		utils.Spacer(),
	}...)

	actions := []map[string]any{}
	if numFeedbackItems > 0 {
		actions = append(actions,
			map[string]any{
				"type": "button",
				"text": map[string]any{
					"type": "plain_text",
					"text": "Stats",
				},
				"action_id": "view-stats",
			},
		)
	}
	if numComments > 0 {
		actions = append(actions,
			map[string]any{
				"type": "button",
				"text": map[string]any{
					"type": "plain_text",
					"text": "Comments",
				},
				"action_id": "view-comments",
			},
		)
	}
	if numFeedbackItems > 0 {
		actions = append(actions,
			map[string]any{
				"type": "button",
				"text": map[string]any{
					"type": "plain_text",
					"text": "Raw Data",
				},
				"action_id": "view-data",
				"confirm": map[string]any{
					"title": map[string]any{
						"type": "plain_text",
						"text": "Fetch Raw Data?",
					},
					"text": map[string]any{
						"type": "plain_text",
						"text": "This action will generate a CSV file of matching raw data. The file will then be sent to you as a direct message.",
					},
					"confirm": map[string]any{
						"type": "plain_text",
						"text": "Export",
					},
					"deny": map[string]any{
						"type": "plain_text",
						"text": "Cancel",
					},
				},
			},
		)
	}
	if len(actions) > 0 {
		blocks = append(blocks, map[string]any{
			"type":     "actions",
			"elements": actions,
		})
	}

	return blocks
}

// Build a static_select element with options and optional initial_option
func staticSelectBlock(actionID, placeholder string, items []string, selected string) map[string]any {
	options := generateOptions(items)
	block := map[string]any{
		"type":        "static_select",
		"placeholder": map[string]any{"type": "plain_text", "text": placeholder},
		"options":     options,
		"action_id":   actionID,
	}
	if selected != "" {
		for _, opt := range options {
			if opt["value"] == selected {
				block["initial_option"] = opt
				break
			}
		}
	}
	return block
}

// Return at least one option for Slack selects
func generateOptions(items []string) []map[string]any {
	if len(items) == 0 {
		return []map[string]any{
			{
				"text":  map[string]any{"type": "plain_text", "text": "No options available"},
				"value": "none",
			},
		}
	}
	options := make([]map[string]any, len(items))
	for i, item := range items {
		options[i] = map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": item},
			"value": item,
		}
	}
	return options
}
