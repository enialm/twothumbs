// File: internal/interactions/settings.go

// This file contains the logic for handling settings-related interactions in the Slack App Home.

package interactions

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"twothumbs/internal/integrations"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/home"
	"twothumbs/internal/templates/modals"
	"twothumbs/internal/utils"
)

// Handle the "home-settings" action
func handleTabSettings(ctx *models.InteractionContext, conn *sql.DB) error {
	apiKey, channel, err := queries.GetAPIKeyAndChannel(conn, ctx.Workspace)
	if err != nil {
		return fmt.Errorf("failed to get api key and channel for workspace %s: %w", ctx.Workspace, err)
	}
	blocks := home.SettingsBlocks(channel, apiKey)
	return PublishHomeView(ctx.BotToken, ctx.UserID, blocks)
}

// Handler for linking a workspace using an activation code
func handleLinkWorkspace(ctx *models.InteractionContext, c *gin.Context, payload map[string]any, conn *sql.DB) error {
	actions, ok := payload["actions"].([]any)
	if !ok || len(actions) == 0 {
		log.Printf("invalid actions payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to process request"})
		return nil
	}
	action, ok := actions[0].(map[string]any)
	if !ok {
		log.Printf("invalid action structure: %v", actions[0])
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to process request"})
		return nil
	}
	value, ok := action["value"].(string)
	if !ok || value == "" {
		log.Printf("missing or invalid activation code")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid activation code"})
		return nil
	}

	linked, err := queries.LinkWorkspace(conn, value, ctx.Workspace)
	if err != nil {
		log.Printf("failed to link workspace: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link workspace"})
		return err
	}
	if !linked {
		c.JSON(http.StatusNotFound, gin.H{"error": "activation code not found or already used"})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{"status": "workspace linked"})

	// Publish the settings view asynchronously
	go func() {
		if err := handleTabSettings(ctx, conn); err != nil {
			log.Printf("failed to publish settings view: %v", err)
		}
	}()

	return nil
}

// Handler for setting the digest channel
func handleSetChannel(ctx *models.InteractionContext, conn *sql.DB, payload map[string]any) error {
	actions, ok := payload["actions"].([]any)
	if !ok || len(actions) == 0 {
		return nil
	}
	action, ok := actions[0].(map[string]any)
	if !ok {
		return nil
	}
	channel, ok := action["selected_channel"].(string)
	if !ok || channel == "" {
		return nil
	}
	if err := queries.UpdateSlackChannelForWorkspace(conn, ctx.Workspace, channel); err != nil {
		return err
	}
	return handleTabSettings(ctx, conn)
}

// Handler for clearing the digest channel
func handleClearChannel(ctx *models.InteractionContext, conn *sql.DB) error {
	if err := queries.ClearSlackChannelForWorkspace(conn, ctx.Workspace); err != nil {
		return err
	}
	return handleTabSettings(ctx, conn)
}

// Handler for rotating the API key
func handleRotateApiKey(ctx *models.InteractionContext, conn *sql.DB) error {
	newKey, err := utils.GenerateUniqueApiKey(conn)
	if err != nil {
		return err
	}
	if err := queries.UpdateApiKeyForWorkspace(conn, ctx.Workspace, newKey); err != nil {
		return err
	}
	return handleTabSettings(ctx, conn)
}

// Handler for viewing test data
func handleViewTestData(ctx *models.InteractionContext, conn *sql.DB) error {
	data, err := queries.GetTestFeedbackData(conn, ctx.Workspace)
	if err != nil {
		return err
	}

	modal := modals.TestDataModal(data)
	return integrations.OpenSlackModal(ctx.TriggerID, modal, ctx.BotToken)
}
