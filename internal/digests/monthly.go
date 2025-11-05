// File: internal/digests/monthly.go

// This file contains the logic to generate monthly digests of feedback comments.

package digests

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/digests"
	"twothumbs/internal/utils"
)

func SendMonthlyDigests(
	conn *sql.DB,
	cfg *config.DigestConfig,
) error {
	workspaces, err := queries.GetActiveWorkspacesAndChannels(conn, models.Monthly)
	if err != nil {
		return fmt.Errorf("failed to get active workspaces: %w", err)
	}

	log.Printf("Processing %d workspaces for monthly digests", len(workspaces))

	var messages []models.DigestMessage

	for _, ws := range workspaces {
		blocks, botToken, err := processMonthlyDigest(conn, ws, cfg)
		if err != nil {
			return fmt.Errorf("failed to process monthly digest for workspace %s: %v", ws.Workspace, err)
		}
		if len(blocks) == 0 {
			log.Printf("No digest data for workspace %s", ws.Workspace)
			continue
		}
		messages = append(messages, models.DigestMessage{
			BotToken:  botToken,
			Channel:   ws.Channel,
			Blocks:    blocks,
			Workspace: ws.Workspace,
		})
	}

	// Wait 10 seconds for Slack to process all file uploads
	log.Printf("Waiting 10 seconds for Slack to finish file processing...")
	time.Sleep(10 * time.Second)

	// Send messages
	for _, msg := range messages {
		err := integrations.SendBlockKitMessage(msg.BotToken, msg.Channel, msg.Blocks)
		if err != nil {
			log.Printf("failed to send monthly digest to workspace %s: %v", msg.Workspace, err)
			continue
		} else {
			log.Printf("Sent monthly digest to workspace %s, channel %s", msg.Workspace, msg.Channel)
		}
	}

	return nil
}

func processMonthlyDigest(
	conn *sql.DB,
	ws models.WorkspaceChannel,
	cfg *config.DigestConfig,
) ([]map[string]any, string, error) {
	groups, err := queries.GetFeedbackGroups(conn, ws.Workspace, models.Monthly)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get feedback groups: %w", err)
	}
	if len(groups) == 0 {
		log.Printf("No feedback groups for workspace %s", ws.Workspace)
		return nil, "", nil
	}

	summaries, err := queries.GetSummariesByWorkspace(conn, ws.Workspace, models.Monthly)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get summaries: %w", err)
	}
	if len(summaries) == 0 {
		log.Printf("No summary data for workspace %s", ws.Workspace)
	}

	botToken, err := queries.GetBotTokenForWorkspace(conn, ws.Workspace)
	if err != nil {
		return nil, "", fmt.Errorf("no bot token found for workspace %s: %w", ws.Workspace, err)
	}

	digestBlocks, err := prepareMonthlyDigestData(conn, cfg, groups, summaries, ws.Workspace, botToken, ws.Channel)
	if err != nil {
		return nil, "", err
	}
	if len(digestBlocks) == 0 {
		log.Printf("No digest data for workspace %s", ws.Workspace)
		return nil, botToken, nil
	}

	blocks := digests.BuildMonthlyDigestBlocks(digestBlocks, utils.MonthLabel())
	return blocks, botToken, nil
}

func prepareMonthlyDigestData(
	conn *sql.DB,
	cfg *config.DigestConfig,
	groups []models.FeedbackGroup, // Only origin and category are used
	summaries []models.SummaryRow,
	workspace, botToken, channel string,
) ([]models.MonthlyDigestData, error) {
	type groupKey struct {
		Origin   string
		Category string
	}
	// Group summaries by origin and category
	summaryMap := make(map[groupKey][]models.SummaryRow)
	for _, s := range summaries {
		key := groupKey{Origin: s.Origin, Category: s.Category}
		summaryMap[key] = append(summaryMap[key], s)
	}

	var digestBlocks []models.MonthlyDigestData
	maxRows := cfg.AIDigestInputLimit

	for _, g := range groups {
		key := groupKey{Origin: g.Origin, Category: g.Category}
		groupSummaries := summaryMap[key]
		groupLen := len(groupSummaries)

		if groupLen > maxRows {
			log.Printf("Encountered %d summary rows (while the limit is %d) for workspace %s", groupLen, maxRows, workspace)
			rand.Shuffle(groupLen, func(i, j int) {
				groupSummaries[i], groupSummaries[j] = groupSummaries[j], groupSummaries[i]
			})
			groupSummaries = groupSummaries[:maxRows]
		}

		digest := "No comments to summarize"
		if groupLen > 0 {
			csvData, err := utils.SummariesToCSV(groupSummaries, true)
			if err != nil {
				return nil, fmt.Errorf("failed to generate CSV for workspace %s, origin %s, category %s: %w", workspace, g.Origin, g.Category, err)
			}
			digest, err = integrations.GetAISummary(cfg.AIApiURL, cfg.AIApiKey, cfg.AIModel, cfg.AIMonthlyDigestPrompt, csvData)
			if err != nil {
				return nil, fmt.Errorf("AI job failed for workspace %s, origin %s, category %s: %w", workspace, g.Origin, g.Category, err)
			}
		}

		stats, err := queries.GetMonthlyFeedbackStats(conn, workspace, g.Origin, g.Category)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for workspace %s, origin %s, category %s: %w", workspace, g.Origin, g.Category, err)
		}

		graphURL, err := GenerateAndUploadMonthlyGraph(conn, botToken, channel, workspace, g.Origin, g.Category)
		if err != nil {
			return nil, fmt.Errorf("failed to generate or upload graph for %s/%s of workspace %s: %w", g.Origin, g.Category, workspace, err)
		}

		scoreDelta := "-"
		if stats.PrevNFeedback > 0 {
			scoreDelta = utils.FormatDelta(stats.ThumbsUpPct - stats.PrevThumbsUpPct)
		}
		digestBlocks = append(digestBlocks, models.MonthlyDigestData{
			Origin:          g.Origin,
			Category:        g.Category,
			Score:           utils.RoundFloat(stats.ThumbsUpPct),
			ScoreDelta:      scoreDelta,
			NResponses:      stats.NFeedback,
			NResponsesDelta: utils.FormatDelta(stats.PrevNFeedbackDl),
			NComments:       stats.NComments,
			NCommentsDelta:  utils.FormatDelta(stats.PrevCommentsDl),
			Digest:          digest,
			GraphURL:        graphURL,
		})
	}

	// Sort blocks alphabetically by origin and category
	sort.Slice(digestBlocks, func(i, j int) bool {
		if digestBlocks[i].Origin != digestBlocks[j].Origin {
			return digestBlocks[i].Origin < digestBlocks[j].Origin
		}
		return digestBlocks[i].Category < digestBlocks[j].Category
	})

	return digestBlocks, nil
}

// Generate and upload a monthly graph, returning the Slack file URL
func GenerateAndUploadMonthlyGraph(
	conn *sql.DB,
	botToken, channel, workspace, origin, category string,
) (string, error) {
	// Get plot stats
	stats, err := queries.GetMonthlyDigestPlotStats(conn, workspace, origin, category)
	if err != nil {
		return "", fmt.Errorf("failed to get plot stats: %w", err)
	}
	// Ensure we have enough stats to plot
	if len(stats) < 3 {
		return "", fmt.Errorf("no stats for origin %s and category %s", origin, category)
	}

	// Plot the graph
	filePath := fmt.Sprintf("plot_%d.png", time.Now().UnixNano())
	defer func() {
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to remove plot file %s: %v", filePath, err)
		}
	}()
	if err := integrations.PlotAPlot(stats, filePath); err != nil {
		return "", fmt.Errorf("failed to plot graph: %w", err)
	}

	// Upload the file to Slack
	url, err := integrations.UploadFileToSlack(botToken, channel, filePath, fmt.Sprintf("Quarterly graph for %s", origin), "")
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Slack: %w", err)
	}

	return url, nil
}
