// File: cmd/ingest/main.go

// This program handles Slack integrations and feedback ingestion.

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"twothumbs/internal/api"
	"twothumbs/internal/config"
	"twothumbs/internal/utils"
)

func main() {
	log.SetOutput(os.Stdout)
	cfg := config.LoadIngestConfig()

	// Connect to the database
	conn, err := utils.ConnectToDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Initialize handlers
	slackHandler := api.NewSlackHandler(conn, cfg)
	feedbackHandler := api.NewFeedbackHandler(conn, cfg)

	router := gin.Default()

	// Add rate limiter middleware
	rateLimiter := utils.NewRateLimiter(
		cfg.FeedbackRateLimitRequests,
		cfg.FeedbackRateLimitWindow,
		cfg.FeedbackRateLimitWindow, // Use window as cleanup interval
	)

	// Health check endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Slack oauth endpoint
	router.GET("/slack/oauth/callback", slackHandler.OAuthCallback)

	// Slack events endpoint
	router.POST("/slack/events", slackHandler.EventsHandler)

	// Feedback endpoint with rate limiting
	router.POST("/feedback", rateLimiter.Limit(), feedbackHandler.PostFeedback)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
