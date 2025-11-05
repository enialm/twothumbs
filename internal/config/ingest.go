// File: internal/config/ingest.go

// This file contains the configuration for the ingest service.

package config

import (
	"strconv"
	"twothumbs/internal/utils"
)

const (
	// Feedback payload char limits
	MaxPromptLen   = 128
	MaxOriginLen   = 32
	MaxCategoryLen = 32
	MaxCommentLen  = 256
	MaxUserIDLen   = 64
)

type IngestConfig struct {
	DatabaseURL               string
	SlackAppURI               string
	SlackAppClientId          string
	SlackAppClientSecret      string
	SlackAppRedirectURI       string
	SlackOAuthRedirectURI     string
	SlackSigningSecret        string
	PromptCountLimit          int // prompts per workspace
	MonthlyFeedbackLimit      int // requests
	FeedbackRateLimitRequests int // requests
	FeedbackRateLimitWindow   int // seconds
}

func LoadIngestConfig() *IngestConfig {
	promptCountLimit, err := strconv.Atoi(utils.GetEnv("PROMPT_COUNT_LIMIT"))
	if err != nil {
		panic("Invalid PROMPT_COUNT_LIMIT: must be an integer")
	}
	monthlyLimit, err := strconv.Atoi(utils.GetEnv("MONTHLY_FEEDBACK_LIMIT"))
	if err != nil {
		panic("Invalid MONTHLY_FEEDBACK_LIMIT: must be an integer")
	}
	rateLimitRequests, err := strconv.Atoi(utils.GetEnv("RATE_LIMIT_REQUESTS"))
	if err != nil {
		panic("Invalid RATE_LIMIT_REQUESTS: must be an integer")
	}
	rateLimitWindow, err := strconv.Atoi(utils.GetEnv("RATE_LIMIT_WINDOW_SECS"))
	if err != nil {
		panic("Invalid RATE_LIMIT_WINDOW_SECS: must be an integer")
	}

	cfg := &IngestConfig{
		DatabaseURL:               utils.GetEnv("DATABASE_URL"),
		SlackAppURI:               utils.GetEnv("SLACK_APP_URI"),
		SlackAppClientId:          utils.GetEnv("SLACK_CLIENT_ID"),
		SlackAppClientSecret:      utils.GetEnv("SLACK_CLIENT_SECRET"),
		SlackAppRedirectURI:       utils.GetEnv("SLACK_APP_REDIRECT_URI"),
		SlackOAuthRedirectURI:     utils.GetEnv("SLACK_OAUTH_REDIRECT_URI"),
		SlackSigningSecret:        utils.GetEnv("SLACK_SIGNING_SECRET"),
		PromptCountLimit:          promptCountLimit,
		MonthlyFeedbackLimit:      monthlyLimit,
		FeedbackRateLimitRequests: rateLimitRequests,
		FeedbackRateLimitWindow:   rateLimitWindow,
	}

	return cfg
}
