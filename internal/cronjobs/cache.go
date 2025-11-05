// File: internal/cronjobs/cache.go

// This file contains the logic to cache daily summaries of feedback comments.

package cronjobs

import (
	"database/sql"
	"log"
	"math/rand"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/utils"
)

func RunDailyCacheJob(conn *sql.DB, cfg *config.DigestConfig) error {
	// Truncate the issues table to keep things tidy
	if err := queries.TruncateIssuesTable(conn); err != nil {
		log.Printf("Failed to truncate issues table: %v", err)
		return err
	}

	workspaces, err := queries.GetActiveWorkspaces(conn)
	if err != nil {
		log.Printf("Failed to get active workspaces: %v", err)
		return err
	}
	for _, workspace := range workspaces {
		if err := processWorkspace(conn, cfg, workspace); err != nil {
			log.Printf("Error processing workspace %s: %v", workspace, err)
			return err
		}
	}
	return nil
}

func processWorkspace(conn *sql.DB, cfg *config.DigestConfig, workspace string) error {
	log.Printf("Processing workspace: %s", workspace)
	feedbacks, err := queries.GetCommentsForCaching(conn, workspace)
	if err != nil {
		log.Printf("Failed to get feedback for workspace %s: %v", workspace, err)
		return err
	}
	log.Printf("Fetched %d feedback rows for workspace %s", len(feedbacks), workspace)

	groups := GroupFeedback(feedbacks)
	log.Printf("Formed %d feedback groups for workspace %s", len(groups), workspace)

	for key, group := range groups {
		if err := processFeedbackGroup(conn, cfg, workspace, key, group); err != nil {
			log.Printf("Error processing feedback group for workspace %s (prompt=%q, origin=%q, category=%q): %v", workspace, key.Prompt, key.Origin, key.Category, err)
			return err
		}
	}

	if err := prepareIssueReports(conn, cfg, workspace); err != nil {
		log.Printf("Error preparing issue reports for workspace %s: %v", workspace, err)
		return err
	}

	return nil
}

func GroupFeedback(feedbacks []*models.Feedback) map[models.FeedbackGroup][]*models.Feedback {
	groups := make(map[models.FeedbackGroup][]*models.Feedback)
	for _, f := range feedbacks {
		key := models.FeedbackGroup{Prompt: f.Prompt, Origin: f.Origin, Category: f.Category}
		groups[key] = append(groups[key], f)
	}
	return groups
}

func processFeedbackGroup(conn *sql.DB, cfg *config.DigestConfig, workspace string, key models.FeedbackGroup, group []*models.Feedback) error {
	maxRows := cfg.AICacheInputLimit
	originalLen := len(group)
	if originalLen > maxRows {
		log.Printf("Encountered %d feedback rows (while the limit was %d) for workspace %s", originalLen, maxRows, workspace)
		rand.Shuffle(originalLen, func(i, j int) {
			group[i], group[j] = group[j], group[i]
		})
		group = group[:maxRows]
	}

	csvData, err := utils.FeedbackToCSV(group)
	if err != nil {
		log.Printf("Failed to encode feedback to CSV for workspace %s: %v", workspace, err)
		return err
	}

	summary, err := integrations.GetAISummary(
		cfg.AIApiURL,
		cfg.AIApiKey,
		cfg.AIModel,
		cfg.AICommentCachePrompt,
		csvData,
	)
	if err != nil {
		log.Printf("AI summary failed for workspace %s (prompt=%q, origin=%q, category=%q): %v", workspace, key.Prompt, key.Origin, key.Category, err)
		return err
	}

	if err := queries.InsertSummary(conn, workspace, key.Prompt, key.Origin, key.Category, len(group), summary); err != nil {
		log.Printf("Failed to insert summary for workspace %s (prompt=%q, origin=%q, category=%q): %v", workspace, key.Prompt, key.Origin, key.Category, err)
		return err
	}

	log.Printf("Inserted summary for workspace %s (prompt=%q, origin=%q, category=%q)", workspace, key.Prompt, key.Origin, key.Category)
	return nil
}

func prepareIssueReports(conn *sql.DB, cfg *config.DigestConfig, workspace string) error {
	// Fetch distinct origins from the summaries table
	origins, err := queries.GetDistinctOrigins(conn, workspace)
	if err != nil {
		log.Printf("Failed to fetch origins for workspace %s: %v", workspace, err)
		return err
	}

	for _, origin := range origins {
		log.Printf("Processing summaries for origin: %s", origin)

		// Fetch summaries for the origin
		summaries, err := queries.GetLatestSummaries(conn, workspace)
		if err != nil {
			log.Printf("Failed to fetch summaries for workspace %s and origin %s: %v", workspace, origin, err)
			return err
		}

		if len(summaries) > cfg.AICacheInputLimit {
			log.Printf("Encountered %d summary rows (while the limit is %d) for workspace %s and origin %s", len(summaries), cfg.AICacheInputLimit, workspace, origin)
			summaries = summaries[:cfg.AICacheInputLimit]
		}

		// Convert []*models.SummaryRow to []models.SummaryRow
		summaryValues := make([]models.SummaryRow, len(summaries))
		for i, summary := range summaries {
			summaryValues[i] = *summary
		}

		// Convert summaries to CSV
		csvData, err := utils.SummariesToCSV(summaryValues, true)
		if err != nil {
			log.Printf("Failed to convert summaries to CSV for workspace %s and origin %s: %v", workspace, origin, err)
			return err
		}

		// Send CSV data to AI for issue report generation
		aiReport, err := integrations.GetAISummary(
			cfg.AIApiURL,
			cfg.AIApiKey,
			cfg.AIModel,
			cfg.AIIssueCachePrompt,
			csvData,
		)
		if err != nil {
			log.Printf("AI issue report generation failed for workspace %s and origin %s: %v", workspace, origin, err)
			return err
		}

		// Skip insertion if the AI report is empty
		if aiReport == "" {
			log.Printf("AI report is empty for workspace %s and origin %s, skipping insertion", workspace, origin)
			continue
		}

		// Insert the issue report into the database
		if err := queries.InsertIssueReport(conn, workspace, origin, aiReport); err != nil {
			log.Printf("Failed to insert issue report for workspace %s and origin %s: %v", workspace, origin, err)
			return err
		}

		log.Printf("Successfully inserted issue report for workspace %s and origin %s", workspace, origin)
	}

	return nil
}
