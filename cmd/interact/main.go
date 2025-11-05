// File: cmd/interact/main.go

// This program handles Slack interactions.

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"twothumbs/internal/config"
	"twothumbs/internal/interactions"
	"twothumbs/internal/utils"
)

func main() {
	log.SetOutput(os.Stdout)
	cfg := config.LoadInteractConfig()

	// Connect to the database
	conn, err := utils.ConnectToDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	router := gin.Default()

	// Health check endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Register Slack interaction handler
	router.POST("/slack/interact", slackInteractHandler(cfg, conn))

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Interaction handler for Slack interactions
func slackInteractHandler(cfg *config.InteractConfig, conn *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload map[string]any
		if err := json.Unmarshal([]byte(c.PostForm("payload")), &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			log.Printf("failed to unmarshal payload: %v", err)
			return
		}

		// Detect interaction type
		switch payload["type"] {
		case "view_submission":
			// Modal submission
			log.Printf("processing a modal submission")
			interactions.HandleModalSubmission(c, payload, conn, cfg)
			return

		case "block_actions":
			// Detect source from container type
			container, ok := payload["container"].(map[string]any)
			if ok {
				switch container["type"] {
				case "view":
					log.Printf("processing a view interaction")
					interactions.HandleViewInteraction(c, payload, conn, cfg)
					return
				case "message":
					log.Printf("processing a message interaction")
					interactions.HandleMessageInteraction(c, payload, conn, cfg)
					return
				}
			}
		}

		// Unsupported payload
		log.Printf("unsupported payload: %v", payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported payload"})
	}
}
