// File: internal/interactions/prompts.go

// This file contains the logic for handling Slack App prompts view interactions.

package interactions

import (
	"database/sql"
	"fmt"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/home"
	"twothumbs/internal/templates/modals"
)

// Handle the "home-prompts" action
func HandleTabPrompts(ctx *models.InteractionContext, conn *sql.DB, cfg *config.InteractConfig) error {
	prompts, err := queries.GetPrompts(conn, ctx.Workspace)
	if err != nil {
		return fmt.Errorf("failed to get prompts for workspace %s: %v", ctx.Workspace, err)
	}
	blocks := home.PromptsBlocks(prompts, cfg.PromptCountLimit)

	return PublishHomeView(ctx.BotToken, ctx.UserID, blocks)
}

// Handle the "modal-delete-prompt" action
func handleModalDeletePrompt(ctx *models.InteractionContext, conn *sql.DB, payload map[string]any) error {
	promptID, err := extractActionValueFromPayload(payload)
	if err != nil {
		return err
	}

	// Fetch the prompt data
	prompt, err := queries.GetPrompt(conn, promptID)
	if err != nil {
		return fmt.Errorf("failed to fetch prompt: %w", err)
	}

	// Build the modal for deleting the prompt
	modal := modals.DeletePromptModal(prompt)

	// Open the modal
	if err := integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken); err != nil {
		return fmt.Errorf("failed to open delete prompt modal: %w", err)
	}

	return nil
}

// Helper to extract action value from the payload
func extractActionValueFromPayload(payload map[string]any) (string, error) {
	actions, ok := payload["actions"].([]any)
	if !ok || len(actions) == 0 {
		return "", fmt.Errorf("invalid actions payload")
	}
	action, ok := actions[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid action structure")
	}
	val, ok := action["value"].(string)
	if !ok || val == "" {
		return "", fmt.Errorf("missing action value")
	}
	return val, nil
}
