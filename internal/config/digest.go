// File: internal/config/digest.go

// This file contains the configuration for the digest service.

package config

import (
	"strconv"
	"twothumbs/internal/utils"
)

type DigestConfig struct {
	DatabaseURL             string
	AIApiURL                string
	AIApiKey                string
	AIModel                 string
	AICacheInputLimit       int
	AIDigestInputLimit      int
	AIDailyDigestPrompt     string
	AIWeeklyDigestPrompt    string
	AIMonthlyDigestPrompt   string
	AIQuarterlyDigestPrompt string
	AICommentCachePrompt    string
	AIIssueCachePrompt      string
}

func LoadDigestConfig() *DigestConfig {
	cache_ilim_str := utils.GetEnv("AI_CACHE_INPUT_LIMIT")
	cache_ilim, err := strconv.Atoi(cache_ilim_str)
	if err != nil || cache_ilim <= 0 {
		panic("Invalid AI_CACHE_INPUT_LIMIT: must be a positive integer")
	}
	digest_ilim_str := utils.GetEnv("AI_DIGEST_INPUT_LIMIT")
	digest_ilim, err := strconv.Atoi(digest_ilim_str)
	if err != nil || digest_ilim <= 0 {
		panic("Invalid AI_DIGEST_INPUT_LIMIT: must be a positive integer")
	}
	cfg := &DigestConfig{
		DatabaseURL:             utils.GetEnv("DATABASE_URL"),
		AIApiURL:                utils.GetEnv("AI_API_URL"),
		AIApiKey:                utils.GetEnv("AI_API_KEY"),
		AIModel:                 utils.GetEnv("AI_MODEL"),
		AICacheInputLimit:       cache_ilim,
		AIDigestInputLimit:      digest_ilim,
		AIDailyDigestPrompt:     utils.GetEnv("AI_DAILY_DIGEST_PROMPT"),
		AIWeeklyDigestPrompt:    utils.GetEnv("AI_WEEKLY_DIGEST_PROMPT"),
		AIMonthlyDigestPrompt:   utils.GetEnv("AI_MONTHLY_DIGEST_PROMPT"),
		AIQuarterlyDigestPrompt: utils.GetEnv("AI_QUARTERLY_DIGEST_PROMPT"),
		AICommentCachePrompt:    utils.GetEnv("AI_COMMENT_CACHE_PROMPT"),
		AIIssueCachePrompt:      utils.GetEnv("AI_ISSUE_CACHE_PROMPT"),
	}

	return cfg
}
