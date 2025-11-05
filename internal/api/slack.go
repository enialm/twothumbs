// File: internal/api/slack.go

// This file provides the Slack integration handler for OAuth callbacks and event handling.
// It handles the OAuth flow, verifies requests, and processes events like app uninstallation and token revocation.
// It should also house the logic to handle incoming Slack App commands.

package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"twothumbs/internal/config"
	"twothumbs/internal/interactions"
	"twothumbs/internal/models"
	"twothumbs/internal/queries"
	"twothumbs/internal/templates/home"
)

type SlackHandler struct {
	DB     *sql.DB
	Config *config.IngestConfig
}

func NewSlackHandler(conn *sql.DB, cfg *config.IngestConfig) *SlackHandler {
	return &SlackHandler{DB: conn, Config: cfg}
}

func (h *SlackHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}

	clientID := h.Config.SlackAppClientId
	clientSecret := h.Config.SlackAppClientSecret
	oAuthRedirectURI := h.Config.SlackOAuthRedirectURI

	tokenResp, err := exchangeCodeForToken(code, clientID, clientSecret, oAuthRedirectURI)
	if err != nil {
		log.Printf("error exchanging code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	inst := &models.Installation{
		SlackWorkspace: tokenResp.Team.ID,
		SlackToken:     tokenResp.AccessToken,
	}

	err = queries.SaveInstallation(h.DB, inst)
	if err != nil {
		log.Printf("error saving installation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save installation"})
		return
	}

	c.Redirect(http.StatusFound, h.Config.SlackAppRedirectURI)
}

func exchangeCodeForToken(code, clientID, clientSecret, redirectURI string) (*models.SlackOAuthResponse, error) {
	form := url.Values{}
	form.Add("code", code)
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("redirect_uri", redirectURI)

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var oauthResp models.SlackOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return nil, err
	}

	if !oauthResp.OK {
		return nil, errors.New("slack oauth error: " + oauthResp.Error)
	}

	return &oauthResp, nil
}

func (h *SlackHandler) EventsHandler(c *gin.Context) {
	signingSecret := h.Config.SlackSigningSecret
	if signingSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server Misconfiguration"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read request body"})
		return
	}
	bodyString := string(bodyBytes)

	timestamp := c.GetHeader("X-Slack-Request-Timestamp")
	slackSignature := c.GetHeader("X-Slack-Signature")
	if !verifySlackRequest(signingSecret, timestamp, bodyString, slackSignature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload["type"] == "url_verification" {
		challenge := payload["challenge"].(string)
		c.JSON(http.StatusOK, gin.H{"challenge": challenge})
		return
	}

	if payload["type"] == "event_callback" {
		event := payload["event"].(map[string]any)
		eventType := event["type"].(string)

		switch eventType {
		case "app_uninstalled":
			workspace := payload["team_id"].(string)
			if err := queries.DeleteInstallation(h.DB, workspace); err != nil {
				log.Printf("error deleting installation for workspace %s: %v", workspace, err)
			} else {
				log.Printf("deleted installation for workspace %s", workspace)
			}

		case "tokens_revoked":
			tokens := event["tokens"].(map[string]any)
			if botTokens, ok := tokens["bot"].([]any); ok {
				for _, botToken := range botTokens {
					tokenStr := botToken.(string)
					if err := queries.DeleteInstallationByToken(h.DB, tokenStr); err != nil {
						log.Printf("error deleting revoked bot token installation: %v", err)
					} else {
						log.Printf("deleted revoked bot token installation")
					}
				}
			}
		case "app_home_opened":
			userID := event["user"].(string)
			workspace := payload["team_id"].(string)
			active, err := queries.IsActiveAccount(h.DB, workspace)
			if err != nil {
				log.Printf("could not check if account is active for workspace %s: %v", workspace, err)
				break
			}
			botToken, err := queries.GetBotTokenForWorkspace(h.DB, workspace)
			if err != nil {
				log.Printf("could not find bot token for workspace %s: %v", workspace, err)
				break
			}
			ctx := &models.InteractionContext{
				BotToken:  botToken,
				UserID:    userID,
				Workspace: workspace,
			}
			go func() {
				if !active {
					if err := interactions.PublishHomeView(botToken, userID, home.WelcomeBlocks()); err != nil {
						log.Printf("failed to publish welcome home tab for user %s: %v", userID, err)
					}
					return
				}
				if err := interactions.HandleTabExplore(ctx, h.DB, false); err != nil {
					log.Printf("failed to publish home tab for user %s: %v", userID, err)
				}
			}()
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func verifySlackRequest(signingSecret, timestamp, body, slackSignature string) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if abs(time.Now().Unix()-ts) > 60*5 {
		return false
	}

	basestring := "v0:" + timestamp + ":" + body
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(basestring))
	expectedSig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(slackSignature))
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
