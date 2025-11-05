// File: internal/api/feedback.go

// This file shall include the endpoint to push feedback to the database.

package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"twothumbs/internal/config"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/utils"

	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct {
	DB     *sql.DB
	Config *config.IngestConfig
}

func NewFeedbackHandler(conn *sql.DB, cfg *config.IngestConfig) *FeedbackHandler {
	return &FeedbackHandler{DB: conn, Config: cfg}
}

func validateFeedbackRequest(req *models.FeedbackRequest) (string, bool) {
	if len(strings.TrimSpace(req.Prompt)) == 0 || len(req.Prompt) > config.MaxPromptLen {
		return fmt.Sprintf("Prompt is required and must be at most %d characters", config.MaxPromptLen), false
	}
	if req.ThumbUp == nil {
		return "ThumbUp is required", false
	}
	if len(strings.TrimSpace(req.Origin)) == 0 || len(req.Origin) > config.MaxOriginLen {
		return fmt.Sprintf("Origin is required and must be at most %d characters", config.MaxOriginLen), false
	}
	if req.InProduction == nil {
		return "InProduction is required", false
	}
	if len(strings.TrimSpace(req.UserID)) == 0 || len(req.UserID) > config.MaxUserIDLen {
		return fmt.Sprintf("UserID is required and must be at most %d characters", config.MaxUserIDLen), false
	}
	if len(strings.TrimSpace(req.Category)) == 0 || len(req.Category) > config.MaxCategoryLen {
		return fmt.Sprintf("Category is required and must be at most %d characters", config.MaxCategoryLen), false
	}
	if len(req.Comment) > config.MaxCommentLen {
		return fmt.Sprintf("Comment must be at most %d characters", config.MaxCommentLen), false
	}
	return "", true
}

// Handler for POST /feedback
func (h *FeedbackHandler) PostFeedback(c *gin.Context) {
	// Authorize via X-API-Key
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		log.Println("Missing X-API-Key header")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	account, err := queries.GetAccount(h.DB, apiKey)
	if err != nil {
		log.Printf("Failed to retrieve account for API key %s: %v", apiKey, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Internal server error"})
		return
	}
	if account == nil || account.SlackWorkspace == nil {
		log.Printf("Invalid API key: %s", apiKey)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Account expiry check
	if account.AccountExpiryDate.Before(time.Now().UTC()) {
		log.Printf("Account expired for workspace %s", *account.SlackWorkspace)
		c.JSON(http.StatusPaymentRequired, models.ErrorResponse{Error: "Account expired"})
		return
	}
	workspace := *account.SlackWorkspace

	// Parse and validate request body
	var req models.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body for workspace %s: %v", workspace, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
		return
	}
	if msg, ok := validateFeedbackRequest(&req); !ok {
		log.Printf("Validation failed for workspace %s: %s", workspace, msg)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: msg})
		return
	}

	// Monthly feedback limit check
	if account.FeedbackCount >= h.Config.MonthlyFeedbackLimit {
		log.Printf("Monthly feedback limit reached for workspace %s", workspace)
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{Error: "Monthly feedback limit reached"})
		return
	}

	// Prompt limit check
	if !h.enforcePromptLimit(workspace, req.Origin, req.Category, req.Prompt) {
		log.Printf("Prompt limit reached for workspace %s", workspace)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Prompt limit reached"})
		return
	}

	// Insert prompt
	if _, err := queries.InsertPrompt(h.DB, workspace, req.Origin, req.Category, req.Prompt); err != nil {
		log.Printf("Failed to insert prompt for workspace %s: %v", workspace, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Internal server error"})
		return
	}

	// Insert feedback
	feedback := models.Feedback{
		SlackWorkspace: workspace,
		Prompt:         req.Prompt,
		ThumbUp:        *req.ThumbUp,
		Comment:        utils.PtrOrNil(req.Comment),
		Origin:         req.Origin,
		Category:       req.Category,
		InProduction:   *req.InProduction,
		UserID:         req.UserID,
	}

	if err := queries.InsertFeedback(h.DB, &feedback); err != nil {
		log.Printf("Failed to insert feedback for workspace %s: %v", workspace, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Internal server error"})
		return
	}

	log.Printf("Feedback submitted successfully for workspace %s", workspace)
	c.JSON(http.StatusCreated, gin.H{"message": "Feedback submitted successfully"})
}

func (h *FeedbackHandler) enforcePromptLimit(workspace, origin, category, prompt string) bool {
	count, exists, err := queries.GetPromptCountAndExists(h.DB, workspace, origin, category, prompt)
	if err != nil {
		log.Printf("Failed to check prompt count/existence for workspace %s: %v", workspace, err)
		return false
	}

	// Allow if prompt exists, or if under limit
	return exists || count < h.Config.PromptCountLimit
}
