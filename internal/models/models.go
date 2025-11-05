// File: internal/models/models.go

// This file contains the data models used in the application, reflecting the database schema.

package models

import "time"

type Installation struct {
	CreatedAt      time.Time `db:"created_at"`
	SlackWorkspace string    `db:"slack_workspace"`
	SlackToken     string    `db:"bot_token"`
}

type Account struct {
	AccountID         string    `db:"account_id"`
	CreatedAt         time.Time `db:"created_at"`
	AccountExpiryDate time.Time `db:"account_expiry_date"`
	ActivationCode    string    `db:"activation_code"`
	APIKey            string    `db:"api_key"`
	SlackWorkspace    *string   `db:"slack_workspace"`
	SlackChannel      *string   `db:"slack_channel"`
	FeedbackCount     int       `db:"feedback_count"`
}

type InteractionContext struct {
	TriggerID string
	Workspace string
	BotToken  string
	UserID    string
}

type StatsFilterValues struct {
	Origin   string
	Category string
	Prompt   string
	DateFrom string
	DateTo   string
	Thumb    string // "Up", "Down", or ""
}

type Prompt struct {
	ID             int64  `db:"id"`
	SlackWorkspace string `db:"slack_workspace"`
	Origin         string `db:"origin"`
	Category       string `db:"category"`
	Prompt         string `db:"prompt"`
}

type Feedback struct {
	ID             int64     `db:"id"`
	CreatedAt      time.Time `db:"created_at"`
	SlackWorkspace string    `db:"slack_workspace"`
	Prompt         string    `db:"prompt"`
	ThumbUp        bool      `db:"thumb_up"`
	Comment        *string   `db:"comment"`
	Origin         string    `db:"origin"`
	Category       string    `db:"category"`
	InProduction   bool      `db:"in_production"`
	UserID         string    `db:"user_id"`
}

type FeedbackRequest struct {
	Prompt       string `json:"prompt" binding:"required"`
	ThumbUp      *bool  `json:"thumb_up" binding:"required"`
	Comment      string `json:"comment"`
	Origin       string `json:"origin" binding:"required"`
	Category     string `json:"category" binding:"required"`
	InProduction *bool  `json:"in_production" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
}

type DigestRange string

const (
	Daily     DigestRange = "1 day"
	Weekly    DigestRange = "7 days"
	Monthly   DigestRange = "1 month"
	Quarterly DigestRange = "3 months"
	Last7d    DigestRange = "7 days"
	Last30d   DigestRange = "30 days"
)

type FeedbackGroup struct {
	Prompt   string
	Origin   string
	Category string
}

type WorkspaceChannel struct {
	Workspace string
	Channel   string
}

type SummaryRow struct {
	SummaryDate time.Time
	Origin      string
	Category    string
	Prompt      string
	NComments   int
	Summary     string
}

type GroupedSummaries struct {
	Workspace string
	Origin    string
	Summaries []SummaryRow
}

type IssueReport struct {
	SlackWorkspace string
	Origin         string
	Report         string
}

type FeedbackStats struct {
	ThumbsUpPct     float64
	PrevThumbsUpPct float64
	NFeedback       int
	PrevNFeedback   int
	PrevNFeedbackDl float64
	NComments       int
	PrevCommentsDl  float64
}

type DigestMessage struct {
	BotToken  string
	Channel   string
	Blocks    []map[string]any
	Workspace string
}

type DailyDigestData struct {
	Origin    string
	NComments int
	Digest    string
}

type WeeklyDigestData struct {
	Origin          string
	Category        string
	Prompt          string
	Score           int
	ScoreDelta      string
	NResponses      int
	NResponsesDelta string
	NComments       int
	NCommentsDelta  string
	Digest          string
}

type MonthlyDigestData struct {
	Origin          string
	Category        string
	Score           int
	ScoreDelta      string
	NResponses      int
	NResponsesDelta string
	NComments       int
	NCommentsDelta  string
	Digest          string
	GraphURL        string
}

type QuarterlyDigestData struct {
	Origin   string
	Digest   string
	GraphURL string
}

type PlotStats struct {
	Month       time.Time
	ThumbsUpPct float64
	NFeedback   int
	NComments   int
}

type ExploreCommentsResult struct {
	Origin    string
	Category  string
	Prompt    string
	Comment   string
	Timestamp time.Time
}

type StatsData struct {
	Origin          string
	Category        string
	Prompt          string
	Score           int
	ScoreDelta      string
	NResponses      int
	NResponsesDelta string
	NComments       int
	NCommentsDelta  string
}

type RawData struct {
	CreatedAt time.Time `db:"created_at"`
	Origin    string    `db:"origin"`
	Category  string    `db:"category"`
	Prompt    string    `db:"prompt"`
	ThumbUp   bool      `db:"thumb_up"`
	Comment   *string   `db:"comment"`
	UserID    string    `db:"user_id"`
}

type UserFilterCacheEntry struct {
	Values    StatsFilterValues
	Timestamp time.Time
}

type EmailSender struct {
	SMTPHost string
	SMTPPort string
	Username string
	Password string
	From     string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SlackOAuthResponse struct {
	OK          bool   `json:"ok"`
	AccessToken string `json:"access_token"`
	Team        struct {
		ID string `json:"id"`
	} `json:"team"`
	Error            string `json:"error"`
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata,omitempty"`
}

type SlackUploadURLResponse struct {
	OK               bool   `json:"ok"`
	UploadURL        string `json:"upload_url"`
	FileID           string `json:"file_id"`
	Error            string `json:"error,omitempty"`
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata,omitempty"`
}

type SlackCompleteUploadResponse struct {
	OK    bool `json:"ok"`
	Files []struct {
		ID                 string `json:"id"`
		Name               string `json:"name"`
		URLPrivate         string `json:"url_private"`
		URLPrivateDownload string `json:"url_private_download"`
		Permalink          string `json:"permalink"`
		PermalinkPublic    string `json:"permalink_public"`
	} `json:"files"`
	Error            string `json:"error,omitempty"`
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata,omitempty"`
}
