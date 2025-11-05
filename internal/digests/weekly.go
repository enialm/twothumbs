// File: internal/digests/weekly.go

// This file contains the logic to generate weekly digests

package digests

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/digests"
	"twothumbs/internal/utils"
)

func SendWeeklyDigests(
	conn *sql.DB,
	cfg *config.DigestConfig,
) error {
	workspaces, err := queries.GetActiveWorkspacesAndChannels(conn, models.Weekly)
	if err != nil {
		return fmt.Errorf("failed to get active workspaces: %w", err)
	}

	log.Printf("Processing %d workspaces for weekly digests", len(workspaces))

	var messages []models.DigestMessage

	for _, ws := range workspaces {
		blocks, botToken, err := processWeeklyDigest(conn, ws, cfg)
		if err != nil {
			return fmt.Errorf("failed to process weekly digest for workspace %s: %v", ws.Workspace, err)
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

	// Send messages
	for _, msg := range messages {
		err := integrations.SendBlockKitMessage(msg.BotToken, msg.Channel, msg.Blocks)
		if err != nil {
			return fmt.Errorf("failed to send weekly digest to workspace %s: %v", msg.Workspace, err)
		} else {
			log.Printf("Sent weekly digest to workspace %s, channel %s", msg.Workspace, msg.Channel)
		}
	}

	return nil
}

func processWeeklyDigest(
	conn *sql.DB,
	ws models.WorkspaceChannel,
	cfg *config.DigestConfig,
) ([]map[string]any, string, error) {
	groups, err := queries.GetFeedbackGroups(conn, ws.Workspace, models.Weekly)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get feedback groups: %w", err)
	}
	if len(groups) == 0 {
		log.Printf("No feedback groups for workspace %s", ws.Workspace)
		return nil, "", nil
	}

	summaries, err := queries.GetSummariesByWorkspace(conn, ws.Workspace, models.Weekly)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get summaries: %w", err)
	}
	if len(summaries) == 0 {
		log.Printf("No summary data for workspace %s", ws.Workspace)
	}

	digestBlocks, err := prepareWeeklyDigestData(conn, cfg, groups, summaries, ws.Workspace)
	if err != nil {
		return nil, "", err
	}
	if len(digestBlocks) == 0 {
		log.Printf("No digest data for workspace %s", ws.Workspace)
		return nil, "", nil
	}

	blocks := digests.BuildWeeklyDigestBlocks(digestBlocks)
	botToken, err := queries.GetBotTokenForWorkspace(conn, ws.Workspace)
	if err != nil {
		return nil, "", fmt.Errorf("no bot token found for workspace %s: %w", ws.Workspace, err)
	}
	return blocks, botToken, nil
}

func prepareWeeklyDigestData(
	conn *sql.DB,
	cfg *config.DigestConfig,
	groups []models.FeedbackGroup,
	summaries []models.SummaryRow,
	workspace string,
) ([]models.WeeklyDigestData, error) {
	// Group summaries by origin, category, and prompt
	summaryMap := make(map[models.FeedbackGroup][]models.SummaryRow)
	for _, s := range summaries {
		key := models.FeedbackGroup{
			Origin:   s.Origin,
			Category: s.Category,
			Prompt:   s.Prompt,
		}
		summaryMap[key] = append(summaryMap[key], s)
	}

	var digestBlocks []models.WeeklyDigestData
	maxRows := cfg.AIDigestInputLimit

	for _, g := range groups {
		key := models.FeedbackGroup{
			Origin:   g.Origin,
			Category: g.Category,
			Prompt:   g.Prompt,
		}
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
			csvData, err := utils.SummariesToCSV(groupSummaries, false)
			if err != nil {
				return nil, fmt.Errorf("failed to generate CSV for workspace %s, origin %s, category %s, prompt %s: %w", workspace, g.Origin, g.Category, g.Prompt, err)
			}
			digest, err = integrations.GetAISummary(cfg.AIApiURL, cfg.AIApiKey, cfg.AIModel, cfg.AIWeeklyDigestPrompt, csvData)
			if err != nil {
				return nil, fmt.Errorf("AI job failed for workspace %s, origin %s, category %s, prompt %s: %w", workspace, g.Origin, g.Category, g.Prompt, err)
			}
		}

		stats, err := queries.GetWeeklyFeedbackStats(conn, workspace, g.Origin, g.Category, g.Prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for workspace %s, origin %s, category %s, prompt %s: %w", workspace, g.Origin, g.Category, g.Prompt, err)
		}

		scoreDelta := "-"
		if stats.PrevNFeedback > 0 {
			scoreDelta = utils.FormatDelta(stats.ThumbsUpPct - stats.PrevThumbsUpPct)
		}
		digestBlocks = append(digestBlocks, models.WeeklyDigestData{
			Origin:          g.Origin,
			Category:        g.Category,
			Prompt:          g.Prompt,
			Score:           utils.RoundFloat(stats.ThumbsUpPct),
			ScoreDelta:      scoreDelta,
			NResponses:      stats.NFeedback,
			NResponsesDelta: utils.FormatDelta(stats.PrevNFeedbackDl),
			NComments:       stats.NComments,
			NCommentsDelta:  utils.FormatDelta(stats.PrevCommentsDl),
			Digest:          digest,
		})
	}

	// Sort blocks alphabetically by origin, category, and prompt
	sort.Slice(digestBlocks, func(i, j int) bool {
		if digestBlocks[i].Origin != digestBlocks[j].Origin {
			return digestBlocks[i].Origin < digestBlocks[j].Origin
		}
		if digestBlocks[i].Category != digestBlocks[j].Category {
			return digestBlocks[i].Category < digestBlocks[j].Category
		}
		return digestBlocks[i].Prompt < digestBlocks[j].Prompt
	})

	return digestBlocks, nil
}
