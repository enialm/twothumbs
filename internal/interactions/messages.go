// File: internal/interactions/messages.go

// This file contains the logic for handling Slack App interactions in messages.

package interactions

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"twothumbs/internal/config"
	"twothumbs/internal/integrations"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/modals"
)

// Process interactions originating from Slack messages
func HandleMessageInteraction(c *gin.Context, payload map[string]any, conn *sql.DB, cfg *config.InteractConfig) {
	actions, ok := payload["actions"].([]any)
	if !ok || len(actions) == 0 {
		log.Printf("invalid or missing actions in payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actions"})
		return
	}

	action := actions[0].(map[string]any)
	actionID, ok := action["action_id"].(string)
	if !ok || actionID == "" {
		log.Printf("missing or invalid action_id in payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action_id"})
		return
	}

	log.Printf("processing message action %s", actionID)

	// Extract the trigger_id
	triggerID, ok := payload["trigger_id"].(string)
	if !ok || triggerID == "" {
		log.Printf("missing or invalid trigger_id in payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid trigger_id"})
		return
	}

	// Extract workspace and bot token
	team, ok := payload["team"].(map[string]any)
	if !ok {
		log.Printf("missing or invalid team in payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid team"})
		return
	}
	workspace := team["id"].(string)

	botToken, err := queries.GetBotTokenForWorkspace(conn, workspace)
	if err != nil || botToken == "" {
		log.Printf("could not find bot token for workspace %s: %v", workspace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find bot token for workspace"})
		return
	}

	// Dispatch to the appropriate handler
	var modal map[string]any
	switch actionID {
	case "explore_feedback":
		modal = modals.ExploreModal()
	case "report_issue":
		modal = modals.ReportModal()
	default:
		log.Printf("unknown message action ID: %s", actionID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown action"})
		return
	}

	// Open the modal using Slack's views.open API
	err = integrations.OpenSlackModal(triggerID, modal, botToken)
	if err != nil {
		log.Printf("failed to handle message action %s: %v", actionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "action handled"})
}
