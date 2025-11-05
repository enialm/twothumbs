// File: internal/interactions/explore.go

// This file contains the logic for handling Slack App home view interactions.

package interactions

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"slices"
	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/home"
	"twothumbs/internal/templates/modals"
	"twothumbs/internal/utils"
)

var userFilterCache sync.Map

// Handle the "home-explore" action
func HandleTabExplore(ctx *models.InteractionContext, conn *sql.DB, expanded bool) error {
	cacheKey := fmt.Sprintf("%s:%s", ctx.UserID, ctx.Workspace)
	var values models.StatsFilterValues
	if v, ok := userFilterCache.Load(cacheKey); ok {
		entry := v.(models.UserFilterCacheEntry)
		if time.Since(entry.Timestamp) < 30*time.Minute {
			values = entry.Values
		} else {
			values = models.StatsFilterValues{
				DateFrom: time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02"),
				DateTo:   time.Now().UTC().Format("2006-01-02"),
			}
		}
	} else {
		values = models.StatsFilterValues{
			DateFrom: time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02"),
			DateTo:   time.Now().UTC().Format("2006-01-02"),
		}
	}

	from := utils.ParseDate(values.DateFrom)
	to := utils.ParseDate(values.DateTo)
	dateInitial := utils.IsDefaultDateRange(from, to)

	origins, categories, prompts, feedbackCount, commentCount, err := queries.GetHomeTabData(
		conn, ctx.Workspace,
		from, to.AddDate(0, 0, 1), // adding one day to get data for the full day
		values.Origin,
		values.Category,
		values.Prompt,
		values.Thumb,
	)
	if err != nil {
		return err
	}

	// Reset any selected value that is no longer valid
	selectedOrigin := ""
	if values.Origin != "" && slices.Contains(origins, values.Origin) {
		selectedOrigin = values.Origin
	}
	selectedCategory := ""
	if values.Category != "" && slices.Contains(categories, values.Category) {
		selectedCategory = values.Category
	}
	selectedPrompt := ""
	if values.Prompt != "" && slices.Contains(prompts, values.Prompt) {
		selectedPrompt = values.Prompt
	}
	selectedThumb := values.Thumb

	// Cache the filter values for this user
	cacheKey = fmt.Sprintf("%s:%s", ctx.UserID, ctx.Workspace)
	userFilterCache.Store(cacheKey, models.UserFilterCacheEntry{
		Values:    values,
		Timestamp: time.Now().UTC(),
	})

	blocks := home.ExploreBlocks(
		origins,
		categories,
		prompts,
		from,
		to,
		dateInitial,
		feedbackCount,
		commentCount,
		selectedOrigin,
		selectedCategory,
		selectedPrompt,
		selectedThumb,
		expanded,
	)

	return PublishHomeView(ctx.BotToken, ctx.UserID, blocks)
}

func handleStatsFilterAction(ctx *models.InteractionContext, conn *sql.DB, payload map[string]any) error {
	values := extractStatsFilterValues(payload)

	if values.DateFrom == "" {
		values.DateFrom = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if values.DateTo == "" {
		values.DateTo = time.Now().UTC().Format("2006-01-02")
	}

	from := utils.ParseDate(values.DateFrom)
	to := utils.ParseDate(values.DateTo)
	dateInitial := utils.IsDefaultDateRange(from, to)

	originsOut, categoriesOut, promptsOut, feedbackCount, commentCount, err := queries.GetHomeTabData(
		conn, ctx.Workspace,
		from, to.AddDate(0, 0, 1),
		values.Origin,
		values.Category,
		values.Prompt,
		values.Thumb,
	)
	if err != nil {
		return err
	}

	// Reset any selected value that is no longer valid
	selectedOrigin := ""
	if values.Origin != "" && slices.Contains(originsOut, values.Origin) {
		selectedOrigin = values.Origin
	}
	selectedCategory := ""
	if values.Category != "" && slices.Contains(categoriesOut, values.Category) {
		selectedCategory = values.Category
	}
	selectedPrompt := ""
	if values.Prompt != "" && slices.Contains(promptsOut, values.Prompt) {
		selectedPrompt = values.Prompt
	}
	selectedThumb := values.Thumb

	// Cache the filter values for this user
	cacheKey := fmt.Sprintf("%s:%s", ctx.UserID, ctx.Workspace)
	userFilterCache.Store(cacheKey, models.UserFilterCacheEntry{
		Values:    values,
		Timestamp: time.Now().UTC(),
	})

	blocks := home.ExploreBlocks(
		originsOut,
		categoriesOut,
		promptsOut,
		from,
		to,
		dateInitial,
		feedbackCount,
		commentCount,
		selectedOrigin,
		selectedCategory,
		selectedPrompt,
		selectedThumb,
		true,
	)

	return PublishHomeView(ctx.BotToken, ctx.UserID, blocks)
}

func extractStatsFilterValues(payload map[string]any) models.StatsFilterValues {
	values := models.StatsFilterValues{}

	view, _ := payload["view"].(map[string]any)
	state, _ := view["state"].(map[string]any)
	vals, _ := state["values"].(map[string]any)

	for _, block := range vals {
		blockMap, ok := block.(map[string]any)
		if !ok {
			continue
		}
		for actionID, action := range blockMap {
			actionMap, ok := action.(map[string]any)
			if !ok {
				continue
			}
			switch actionID {
			case "select-origin":
				if sel, ok := actionMap["selected_option"].(map[string]any); ok && sel != nil {
					if v, ok := sel["value"].(string); ok {
						values.Origin = v
					}
				}
			case "select-category":
				if sel, ok := actionMap["selected_option"].(map[string]any); ok && sel != nil {
					if v, ok := sel["value"].(string); ok {
						values.Category = v
					}
				}
			case "select-prompt":
				if sel, ok := actionMap["selected_option"].(map[string]any); ok && sel != nil {
					if v, ok := sel["value"].(string); ok {
						values.Prompt = v
					}
				}
			case "select-thumb":
				if sel, ok := actionMap["selected_option"].(map[string]any); ok && sel != nil {
					if v, ok := sel["value"].(string); ok {
						values.Thumb = v
					}
				}
			case "stats-date-from":
				if v, ok := actionMap["selected_date"].(string); ok {
					values.DateFrom = v
				}
			case "stats-date-to":
				if v, ok := actionMap["selected_date"].(string); ok {
					values.DateTo = v
				}
			}
		}
	}
	return values
}

func handleClearFilter(ctx *models.InteractionContext, conn *sql.DB, payload map[string]any, action string) error {
	values := extractStatsFilterValues(payload)

	switch action {
	case "clear-origin":
		values.Origin = ""
	case "clear-category":
		values.Category = ""
	case "clear-prompt":
		values.Prompt = ""
	case "clear-thumb":
		values.Thumb = ""
	case "clear-dates":
		values.DateFrom = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
		values.DateTo = time.Now().UTC().Format("2006-01-02")
	}

	from := utils.ParseDate(values.DateFrom)
	to := utils.ParseDate(values.DateTo)
	dateInitial := utils.IsDefaultDateRange(from, to)

	origins, categories, prompts, feedbackCount, commentCount, err := queries.GetHomeTabData(
		conn, ctx.Workspace,
		from, to.AddDate(0, 0, 1),
		values.Origin,
		values.Category,
		values.Prompt,
		values.Thumb,
	)
	if err != nil {
		return err
	}

	// Reset any selected value that is no longer valid
	selectedOrigin := ""
	if values.Origin != "" && slices.Contains(origins, values.Origin) {
		selectedOrigin = values.Origin
	}
	selectedCategory := ""
	if values.Category != "" && slices.Contains(categories, values.Category) {
		selectedCategory = values.Category
	}
	selectedPrompt := ""
	if values.Prompt != "" && slices.Contains(prompts, values.Prompt) {
		selectedPrompt = values.Prompt
	}
	selectedThumb := values.Thumb

	// Cache the filter values for this user
	cacheKey := fmt.Sprintf("%s:%s", ctx.UserID, ctx.Workspace)
	userFilterCache.Store(cacheKey, models.UserFilterCacheEntry{
		Values:    values,
		Timestamp: time.Now().UTC(),
	})

	blocks := home.ExploreBlocks(
		origins,
		categories,
		prompts,
		from,
		to,
		dateInitial,
		feedbackCount,
		commentCount,
		selectedOrigin,
		selectedCategory,
		selectedPrompt,
		selectedThumb,
		true,
	)

	return PublishHomeView(ctx.BotToken, ctx.UserID, blocks)
}

func handleStatsViewAction(ctx *models.InteractionContext, conn *sql.DB, payload map[string]any) error {
	values := extractStatsFilterValues(payload)
	loc, _ := time.LoadLocation("UTC")
	from, _ := time.ParseInLocation("2006-01-02", values.DateFrom, loc)
	to, _ := time.ParseInLocation("2006-01-02", values.DateTo, loc)

	periodDays := int(to.AddDate(0, 0, 1).Sub(from).Hours() / 24)
	prevTo := from.AddDate(0, 0, -1)
	prevFrom := prevTo.AddDate(0, 0, -periodDays+1)

	groups, err := queries.GetFeedbackGroupsWithFilters(
		conn, ctx.Workspace, from, to.AddDate(0, 0, 1),
		values.Origin,
		values.Category,
		values.Prompt,
		values.Thumb,
	)
	if err != nil {
		return err
	}

	stats := FetchStatsWithRange(
		conn, ctx.Workspace, groups, from, to.AddDate(0, 0, 1), prevFrom, prevTo, values.Thumb, queries.GetFilteredStats,
	)

	modal := modals.StatsModal("Stats 📊", stats)
	return integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken)
}

func handleCommentsViewAction(ctx *models.InteractionContext, conn *sql.DB, cfg *config.InteractConfig, payload map[string]any) error {
	values := extractStatsFilterValues(payload)
	loc, _ := time.LoadLocation("UTC")
	from, _ := time.ParseInLocation("2006-01-02", values.DateFrom, loc)
	to, _ := time.ParseInLocation("2006-01-02", values.DateTo, loc)

	results, err := queries.GetCommentsWithFilters(
		conn, ctx.Workspace, from, to.AddDate(0, 0, 1),
		values.Origin,
		values.Category,
		values.Prompt,
		values.Thumb,
		cfg.NComments,
	)
	if err != nil {
		return err
	}

	modal := modals.ExploreCommentsModal("Comments  💬", cfg, results)
	return integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken)
}

func handleRawDataViewAction(ctx *models.InteractionContext, conn *sql.DB, payload map[string]any) error {
	values := extractStatsFilterValues(payload)
	loc, _ := time.LoadLocation("UTC")
	from, _ := time.ParseInLocation("2006-01-02", values.DateFrom, loc)
	to, _ := time.ParseInLocation("2006-01-02", values.DateTo, loc)

	// Fetch raw data
	rawData, err := queries.GetRawFeedbackData(
		conn, ctx.Workspace, from, to.AddDate(0, 0, 1),
		values.Origin,
		values.Category,
		values.Prompt,
		values.Thumb,
	)
	if err != nil {
		return err
	}

	// Generate CSV
	csvStr, err := utils.RawDataToCSV(rawData)
	if err != nil {
		return err
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "feedback-*.csv")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(csvStr); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Upload to Slack (send as DM to user)
	appHomeChannel, err := integrations.GetAppHomeChannelID(ctx.BotToken, ctx.UserID)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"Attached is your feedback export from %s to %s.",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	_, err = integrations.UploadFileToSlack(
		ctx.BotToken, appHomeChannel, tmpFile.Name(),
		"Feedback Export",
		msg,
	)
	if err != nil {
		return err
	}

	return nil
}
