// File: internal/interactions/common.go

// This file contains common functions used across Slack App interactions.

package interactions

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/modals"
	"twothumbs/internal/utils"

	"github.com/gin-gonic/gin"
)

// Handle view interactions
func HandleViewInteraction(c *gin.Context, payload map[string]any, conn *sql.DB, cfg *config.InteractConfig) {
	// Extract interaction context
	ctx, err := ExtractInteractionContext(payload, conn)
	if err != nil {
		log.Printf("failed to extract interaction context: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract action ID
	actions, ok := payload["actions"].([]any)
	if !ok || len(actions) == 0 {
		log.Printf("invalid actions payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actions payload"})
		return
	}
	action, ok := actions[0].(map[string]any)
	if !ok {
		log.Printf("invalid action structure: %v", actions[0])
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action structure"})
		return
	}
	actionID, ok := action["action_id"].(string)
	if !ok || actionID == "" {
		log.Printf("missing or invalid action ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid action ID"})
		return
	}

	// Delegate to the appropriate handler based on action ID
	switch actionID {
	// Home / main actions
	case "home-explore":
		err = HandleTabExplore(ctx, conn, false)
	case "home-expand":
		err = HandleTabExplore(ctx, conn, true)
	case "home-prompts":
		err = HandleTabPrompts(ctx, conn, cfg)
	case "home-settings":
		err = handleTabSettings(ctx, conn)
	case "home-support":
		err = integrations.OpenSlackModal(ctx.TriggerID, modals.SupportModal(), ctx.BotToken)

	// Home / quick actions
	case "home-comments":
		err = handleComments(ctx, conn, cfg, false)
	case "home-top-issues":
		err = handleTopIssues(ctx, conn, false)
	case "home-stats-7d":
		err = handleStats("7-Day Stats", models.Last7d, queries.GetLast7DayStats, false)(ctx, conn, cfg)
	case "home-stats-30d":
		err = handleStats("30-Day Stats", models.Last30d, queries.GetLast30DayStats, false)(ctx, conn, cfg)

	// Home / filter actions
	case "select-origin", "select-category", "select-prompt", "select-thumb", "stats-date-from", "stats-date-to":
		err = handleStatsFilterAction(ctx, conn, payload)
	case "clear-origin":
		err = handleClearFilter(ctx, conn, payload, actionID)
	case "clear-category":
		err = handleClearFilter(ctx, conn, payload, actionID)
	case "clear-prompt":
		err = handleClearFilter(ctx, conn, payload, actionID)
	case "clear-thumb":
		err = handleClearFilter(ctx, conn, payload, actionID)
	case "clear-dates":
		err = handleClearFilter(ctx, conn, payload, actionID)

	// Home / query actions
	case "view-stats":
		err = handleStatsViewAction(ctx, conn, payload)
	case "view-comments":
		err = handleCommentsViewAction(ctx, conn, cfg, payload)
	case "view-data":
		err = handleRawDataViewAction(ctx, conn, payload)

	// Home / prompt actions
	case "modal-delete-prompt":
		err = handleModalDeletePrompt(ctx, conn, payload)

	// Explore Feedback modal / quick actions
	case "modal-comments":
		err = handleComments(ctx, conn, cfg, true)
	case "modal-top-issues":
		err = handleTopIssues(ctx, conn, true)
	case "modal-stats-7d":
		err = handleStats("7-Day Stats", models.Last7d, queries.GetLast7DayStats, true)(ctx, conn, cfg)
	case "modal-stats-30d":
		err = handleStats("30-Day Stats", models.Last30d, queries.GetLast30DayStats, true)(ctx, conn, cfg)

	// Settings actions
	case "set-channel":
		err = handleSetChannel(ctx, conn, payload)
	case "clear-channel":
		err = handleClearChannel(ctx, conn)
	case "rotate-api-key":
		err = handleRotateApiKey(ctx, conn)
	case "view-test-data":
		err = handleViewTestData(ctx, conn)

	// Link workspace action
	case "link-workspace":
		handleLinkWorkspace(ctx, c, payload, conn)
		return

	default:
		log.Printf("unknown view action ID: %s", actionID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown action"})
		return
	}

	// Handle errors from the delegated handler
	if err != nil {
		log.Printf("failed to handle action %s: %v", actionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "action handled"})
}

// ExtractInteractionContext extracts common fields from a Slack interaction payload
func ExtractInteractionContext(payload map[string]any, conn *sql.DB) (*models.InteractionContext, error) {
	// Extract trigger_id
	triggerID, ok := payload["trigger_id"].(string)
	if !ok || triggerID == "" {
		return nil, fmt.Errorf("missing or invalid trigger_id")
	}

	// Extract workspace ID
	team, ok := payload["team"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing or invalid team in payload")
	}
	workspace, ok := team["id"].(string)
	if !ok || workspace == "" {
		return nil, fmt.Errorf("missing or invalid workspace ID")
	}

	// Extract user_id
	user, ok := payload["user"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing or invalid user in payload")
	}
	userID, ok := user["id"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("missing or invalid user ID")
	}

	// Fetch bot token for the workspace
	botToken, err := queries.GetBotTokenForWorkspace(conn, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bot token for workspace %s: %w", workspace, err)
	}

	return &models.InteractionContext{
		TriggerID: triggerID,
		Workspace: workspace,
		BotToken:  botToken,
		UserID:    userID,
	}, nil
}

// Publish the App Home view
func PublishHomeView(botToken, userID string, blocks []map[string]any) error {
	view := map[string]any{
		"type":   "home",
		"blocks": blocks,
	}

	payload := map[string]any{
		"user_id": userID,
		"view":    view,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal home view payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/views.publish", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var respBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to decode Slack response: %w", err)
	}
	if ok, exists := respBody["ok"].(bool); !exists || !ok {
		return fmt.Errorf("slack API error: %v", respBody)
	}

	return nil
}

// Process feedback groups and return stats data
func FetchStats(conn *sql.DB, workspace string, groups []models.FeedbackGroup, fetchFunc func(*sql.DB, string, string, string, string) (*models.FeedbackStats, error)) []models.StatsData {
	var stats []models.StatsData
	for _, group := range groups {
		stat, err := fetchFunc(conn, workspace, group.Origin, group.Category, group.Prompt)
		if err != nil {
			log.Printf("failed to get stats for workspace %s, origin %s, category %s, prompt %s: %v",
				workspace, group.Origin, group.Category, group.Prompt, err)
			continue
		}
		if stat == nil {
			log.Printf("no stats found for workspace %s, origin %s, category %s, prompt %s",
				workspace, group.Origin, group.Category, group.Prompt)
			continue
		}
		scoreDelta := "-"
		if stat.PrevNFeedback > 0 {
			scoreDelta = utils.FormatDelta(stat.ThumbsUpPct - stat.PrevThumbsUpPct)
		}
		stats = append(stats, models.StatsData{
			Origin:          group.Origin,
			Category:        group.Category,
			Prompt:          group.Prompt,
			Score:           utils.RoundFloat(stat.ThumbsUpPct),
			ScoreDelta:      scoreDelta,
			NResponses:      stat.NFeedback,
			NResponsesDelta: utils.FormatDelta(stat.PrevNFeedbackDl),
			NComments:       stat.NComments,
			NCommentsDelta:  utils.FormatDelta(stat.PrevCommentsDl),
		})
	}

	// Sort stats by origin and category
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Origin != stats[j].Origin {
			return stats[i].Origin < stats[j].Origin
		}
		return stats[i].Category < stats[j].Category
	})

	return stats
}

func FetchStatsWithRange(
	conn *sql.DB,
	workspace string,
	groups []models.FeedbackGroup,
	from, to, prevFrom, prevTo time.Time,
	thumb string,
	fetchFunc func(*sql.DB, string, string, string, string, time.Time, time.Time, time.Time, time.Time, string) (*models.FeedbackStats, error),
) []models.StatsData {
	var stats []models.StatsData
	for _, group := range groups {
		stat, err := fetchFunc(conn, workspace, group.Origin, group.Category, group.Prompt, from, to, prevFrom, prevTo, thumb)
		if err != nil {
			log.Printf("failed to get stats for workspace %s, origin %s, category %s, prompt %s: %v",
				workspace, group.Origin, group.Category, group.Prompt, err)
			continue
		}
		if stat == nil {
			log.Printf("no stats found for workspace %s, origin %s, category %s, prompt %s",
				workspace, group.Origin, group.Category, group.Prompt)
			continue
		}
		scoreDelta := "-"
		if stat.PrevNFeedback > 0 {
			scoreDelta = utils.FormatDelta(stat.ThumbsUpPct - stat.PrevThumbsUpPct)
		}
		stats = append(stats, models.StatsData{
			Origin:          group.Origin,
			Category:        group.Category,
			Prompt:          group.Prompt,
			Score:           utils.RoundFloat(stat.ThumbsUpPct),
			ScoreDelta:      scoreDelta,
			NResponses:      stat.NFeedback,
			NResponsesDelta: utils.FormatDelta(stat.PrevNFeedbackDl),
			NComments:       stat.NComments,
			NCommentsDelta:  utils.FormatDelta(stat.PrevCommentsDl),
		})
	}

	// Sort stats by origin and category
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Origin != stats[j].Origin {
			return stats[i].Origin < stats[j].Origin
		}
		return stats[i].Category < stats[j].Category
	})

	return stats
}
