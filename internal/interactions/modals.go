// File: internal/interactions/modals.go

// This file contains the logic for handling Slack App modals.

package interactions

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates"
	"twothumbs/internal/templates/modals"
)

// Display the "comments" modal with the latest comments
func handleComments(ctx *models.InteractionContext, conn *sql.DB, cfg *config.InteractConfig, isModal bool) error {
	// Fetch the latest comments
	results, err := queries.GetLatestComments(conn, ctx.Workspace, cfg.NComments)
	if err != nil {
		log.Printf("failed to get latest comments for workspace %s: %v", ctx.Workspace, err)
		return err
	}

	// Construct the modal
	modal := modals.ExploreCommentsModal("Latest Comments  💬", cfg, results)

	// Push or open the modal based on the isModal flag
	if isModal {
		return integrations.PushSlackView(ctx.TriggerID, modal, ctx.BotToken)
	}
	return integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken)
}

// Push the "top issues" modal
func handleTopIssues(ctx *models.InteractionContext, conn *sql.DB, isModal bool) error {
	// Fetch the top issues for the workspace
	issues, err := queries.GetTopIssuesByWorkspace(conn, ctx.Workspace)
	if err != nil {
		log.Printf("failed to get top issues for workspace %s: %v", ctx.Workspace, err)
		return err
	}

	// Construct the modal
	modal := modals.TopIssuesModal(issues)

	// Push or open the modal based on the isModal flag
	if isModal {
		return integrations.PushSlackView(ctx.TriggerID, modal, ctx.BotToken)
	}
	return integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken)
}

// Generic handler for stats modals
func handleStats(
	title string,
	period models.DigestRange,
	fetchFunc func(*sql.DB, string, string, string, string) (*models.FeedbackStats, error),
	isModal bool,
) func(*models.InteractionContext, *sql.DB, *config.InteractConfig) error {
	return func(ctx *models.InteractionContext, conn *sql.DB, cfg *config.InteractConfig) error {
		// Fetch feedback groups for the given period
		groups, err := queries.GetFeedbackGroups(conn, ctx.Workspace, period)
		if err != nil {
			log.Printf("failed to get feedback groups for workspace %s: %v", ctx.Workspace, err)
			return err
		}

		// Fetch stats for each group
		stats := FetchStats(conn, ctx.Workspace, groups, fetchFunc)

		// Construct the modal
		modal := modals.StatsModal(title, stats)

		// Push or open the modal based on the isModal flag
		if isModal {
			return integrations.PushSlackView(ctx.TriggerID, modal, ctx.BotToken)
		}
		return integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken)
	}
}

// Process modal submissions
func HandleModalSubmission(c *gin.Context, payload map[string]any, conn *sql.DB, cfg *config.InteractConfig) {
	view, ok := payload["view"].(map[string]any)
	if !ok {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "missing view in payload"})
		return
	}
	callbackID, _ := view["callback_id"].(string)

	switch callbackID {
	case "contact-support":
		handleContactSupportSubmission(c, payload, cfg)
	case "delete-prompt":
		handleDeletePromptSubmission(c, payload, conn, cfg)
	default:
		c.JSON(http.StatusOK, map[string]any{
			"response_action": "clear",
		})
	}
}

// Handle contact-support modal submission
func handleContactSupportSubmission(c *gin.Context, payload map[string]any, cfg *config.InteractConfig) {
	fields := extractModalSubmissionData(payload)
	workspace := fields["workspace"]
	message := fields["message-text"]
	issue := fields["issue-text"]
	email := fields["email-address"]

	var subject, body string
	if issue != "" {
		subject, body = templates.IssueEmail(workspace, message)
	} else {
		subject, body = templates.ContactSupportEmail(workspace, message)
	}
	emailSender := integrations.NewEmailSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
		cfg.SMTPFrom,
	)

	err := integrations.SendEmail(emailSender, cfg.SMTPFrom, email, subject, body)
	if err != nil {
		log.Printf("failed to send a support/issue email: %v", err)
		updateSlackModal(c, modals.ReportErrorModal(cfg.SMTPFrom))
		return
	}

	if issue != "" {
		updateSlackModal(c, modals.ReportSentModal())
	} else {
		updateSlackModal(c, modals.ContactSupportSentModal())
	}
}

// Handle delete-prompt submission
func handleDeletePromptSubmission(c *gin.Context, payload map[string]any, conn *sql.DB, cfg *config.InteractConfig) {
	ctx, err := ExtractInteractionContext(payload, conn)
	if err != nil {
		log.Printf("failed to extract interaction context: %v", err)
		updateSlackModal(c, modals.DeletePromptErrorModal())
		return
	}

	view := payload["view"].(map[string]any)
	promptID, ok := view["private_metadata"].(string)
	if !ok || promptID == "" {
		log.Printf("missing prompt id in private_metadata")
		updateSlackModal(c, modals.DeletePromptErrorModal())
		return
	}

	fields := extractModalSubmissionData(payload)
	confirm := fields["delete-confirm"]

	if confirm != "YES" {
		updateSlackModal(c, modals.DeletePromptNoYesModal())
		return
	}

	if err := queries.DeletePromptAndFeedback(conn, promptID); err != nil {
		log.Printf("failed to delete prompt and feedback: %v", err)
		updateSlackModal(c, modals.DeletePromptErrorModal())
		return
	}

	updateSlackModal(c, modals.DeletePromptSuccessModal())

	go func() {
		_ = HandleTabPrompts(ctx, conn, cfg)
	}()
}

// Extract data from a modal submission payload
func extractModalSubmissionData(payload map[string]any) map[string]string {
	result := make(map[string]string)

	view, ok := payload["view"].(map[string]any)
	if !ok {
		return result
	}
	state, ok := view["state"].(map[string]any)
	if !ok {
		return result
	}
	values, ok := state["values"].(map[string]any)
	if !ok {
		return result
	}

	for _, block := range values {
		blockMap, ok := block.(map[string]any)
		if !ok {
			continue
		}
		for actionID, action := range blockMap {
			actionMap, ok := action.(map[string]any)
			if !ok {
				continue
			}
			if val, exists := actionMap["value"].(string); exists {
				result[actionID] = val
			}
		}
	}

	return result
}

// Update the current Slack modal using response_action "update"
func updateSlackModal(c *gin.Context, modal map[string]any) {
	c.JSON(http.StatusOK, map[string]any{
		"response_action": "update",
		"view":            modal,
	})
}
