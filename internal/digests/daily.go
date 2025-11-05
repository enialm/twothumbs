// File: internal/digests/daily.go

// This file contains the logic to generate daily digests

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

func SendDailyDigests(
	conn *sql.DB,
	cfg *config.DigestConfig,
) error {
	workspaces, err := queries.GetActiveWorkspacesAndChannels(conn, models.Daily)
	if err != nil {
		return fmt.Errorf("failed to get active workspaces: %w", err)
	}

	log.Printf("Processing %d workspaces for daily digests", len(workspaces))

	var messages []models.DigestMessage

	for _, ws := range workspaces {
		blocks, botToken, err := processDailyDigest(conn, ws, cfg)
		if err != nil {
			return fmt.Errorf("failed to process daily digest for workspace %s: %v", ws.Workspace, err)
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
			log.Printf("failed to send daily digest to workspace %s: %v", msg.Workspace, err)
			continue
		} else {
			log.Printf("Sent daily digest to workspace %s, channel %s", msg.Workspace, msg.Channel)
		}
	}

	return nil
}

func processDailyDigest(
	conn *sql.DB,
	ws models.WorkspaceChannel,
	cfg *config.DigestConfig,
) ([]map[string]any, string, error) {
	summaries, err := queries.GetSummariesByWorkspace(conn, ws.Workspace, models.Daily)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get summaries: %w", err)
	}
	if len(summaries) == 0 {
		log.Printf("No summary data for workspace %s", ws.Workspace)
	}

	digestBlocks, err := prepareDailyDigestData(cfg, summaries, ws.Workspace)
	if err != nil {
		return nil, "", err
	}
	if len(digestBlocks) == 0 {
		log.Printf("No digest data for workspace %s", ws.Workspace)
		return nil, "", nil
	}

	blocks := digests.BuildDailyDigestBlocks(digestBlocks)
	botToken, err := queries.GetBotTokenForWorkspace(conn, ws.Workspace)
	if err != nil {
		return nil, "", fmt.Errorf("no bot token found for workspace %s: %w", ws.Workspace, err)
	}
	return blocks, botToken, nil
}

func prepareDailyDigestData(
	cfg *config.DigestConfig,
	summaries []models.SummaryRow,
	workspace string,
) ([]models.DailyDigestData, error) {
	// Group summaries by origin
	originMap := make(map[string][]models.SummaryRow)
	for _, s := range summaries {
		originMap[s.Origin] = append(originMap[s.Origin], s)
	}

	var digestBlocks []models.DailyDigestData
	maxRows := cfg.AIDigestInputLimit

	for origin, group := range originMap {
		originalLen := len(group)
		if originalLen > maxRows {
			log.Printf("Encountered %d summary rows (while the limit is %d) for workspace %s", originalLen, maxRows, workspace)
			rand.Shuffle(originalLen, func(i, j int) {
				group[i], group[j] = group[j], group[i]
			})
			group = group[:maxRows]
		}

		csvData, err := utils.SummariesToCSV(group, false)
		if err != nil {
			return nil, fmt.Errorf("failed to generate CSV for workspace %s, origin %s: %w", workspace, origin, err)
		}

		digest, err := integrations.GetAISummary(cfg.AIApiURL, cfg.AIApiKey, cfg.AIModel, cfg.AIDailyDigestPrompt, csvData)
		if err != nil {
			return nil, fmt.Errorf("AI job failed for workspace %s, origin %s: %w", workspace, origin, err)
		}

		nComments := 0
		for _, s := range group {
			nComments += s.NComments
		}
		digestBlocks = append(digestBlocks, models.DailyDigestData{
			Origin:    origin,
			NComments: nComments,
			Digest:    digest,
		})
	}

	// Sort blocks alphabetically by origin
	sort.Slice(digestBlocks, func(i, j int) bool {
		return digestBlocks[i].Origin < digestBlocks[j].Origin
	})

	return digestBlocks, nil
}
