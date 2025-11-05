// File: internal/config/interact.go

// This file contains the configuration for the interact service.

package config

import (
	"strconv"
	"twothumbs/internal/utils"
)

type InteractConfig struct {
	DatabaseURL          string
	SMTPHost             string
	SMTPPort             string
	SMTPUser             string
	SMTPPass             string
	SMTPFrom             string
	NComments            int
	PromptCountLimit     int // prompts per workspace
	MonthlyFeedbackLimit int // requests
}

func LoadInteractConfig() *InteractConfig {
	nComments, err := strconv.Atoi(utils.GetEnv("N_COMMENTS"))
	if err != nil {
		panic("Invalid N_COMMENTS: must be an integer")
	}
	promptCountLimit, err := strconv.Atoi(utils.GetEnv("PROMPT_COUNT_LIMIT"))
	if err != nil {
		panic("Invalid PROMPT_COUNT_LIMIT: must be an integer")
	}
	monthlyLimit, err := strconv.Atoi(utils.GetEnv("MONTHLY_FEEDBACK_LIMIT"))
	if err != nil {
		panic("Invalid MONTHLY_FEEDBACK_LIMIT: must be an integer")
	}

	cfg := &InteractConfig{
		DatabaseURL:          utils.GetEnv("DATABASE_URL"),
		SMTPHost:             utils.GetEnv("SMTP_HOST"),
		SMTPPort:             utils.GetEnv("SMTP_PORT"),
		SMTPUser:             utils.GetEnv("SMTP_USERNAME"),
		SMTPPass:             utils.GetEnv("SMTP_PASSWORD"),
		SMTPFrom:             utils.GetEnv("SMTP_FROM"),
		NComments:            nComments,
		PromptCountLimit:     promptCountLimit,
		MonthlyFeedbackLimit: monthlyLimit,
	}

	return cfg
}
