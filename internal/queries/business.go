// File: internal/db/queries/business.go

// This file contains queries for managing account and installation records in the database.

package queries

import (
	"database/sql"

	"twothumbs/internal/models"
)

// *Table: installations*

// Save a Slack installation record
func SaveInstallation(conn *sql.DB, inst *models.Installation) error {
	query := `
	INSERT INTO installations (slack_workspace, bot_token)
	VALUES ($1, $2)
	ON CONFLICT (slack_workspace) 
	DO UPDATE SET bot_token = EXCLUDED.bot_token
	`
	_, err := conn.Exec(query, inst.SlackWorkspace, inst.SlackToken)
	return err
}

// Delete a Slack installation record by workspace
func DeleteInstallation(conn *sql.DB, slackWorkspace string) error {
	query := `DELETE FROM installations WHERE slack_workspace = $1`
	_, err := conn.Exec(query, slackWorkspace)
	return err
}

// Delete a Slack installation record by token
func DeleteInstallationByToken(conn *sql.DB, slackToken string) error {
	query := `DELETE FROM installations WHERE bot_token = $1`
	_, err := conn.Exec(query, slackToken)
	return err
}

// Get the Slack bot token for a given workspace
func GetBotTokenForWorkspace(conn *sql.DB, slackWorkspace string) (string, error) {
	var botToken string
	err := conn.QueryRow(
		`SELECT bot_token FROM installations WHERE slack_workspace = $1 LIMIT 1`,
		slackWorkspace,
	).Scan(&botToken)
	if err != nil {
		return "", err
	}
	return botToken, nil
}

// *Table: accounts*

// Get account information for a given API key
func GetAccount(conn *sql.DB, apiKey string) (*models.Account, error) {
	row := conn.QueryRow(`
        SELECT account_expiry_date, slack_workspace, feedback_count
        FROM accounts
        WHERE api_key = $1
        LIMIT 1
    `, apiKey)

	var cust models.Account
	err := row.Scan(
		&cust.AccountExpiryDate,
		&cust.SlackWorkspace,
		&cust.FeedbackCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cust, nil
}

// Check if a account is active (workspace is linked and account_expiry_date > CURRENT_DATE)
func IsActiveAccount(conn *sql.DB, slackWorkspace string) (bool, error) {
	var active bool
	query := `
        SELECT EXISTS(
            SELECT 1 FROM accounts
            WHERE slack_workspace = $1
              AND account_expiry_date > CURRENT_DATE
			LIMIT 1
        )
    `
	err := conn.QueryRow(query, slackWorkspace).Scan(&active)
	if err != nil {
		return false, err
	}
	return active, nil
}

// Link a workspace to a account using the activation code, returning true if successful
func LinkWorkspace(conn *sql.DB, activationCode, workspace string) (bool, error) {
	query := `
        UPDATE accounts
        SET slack_workspace = $1
        WHERE activation_code = $2
          AND slack_workspace IS NULL
        RETURNING 1
    `
	var updated int
	err := conn.QueryRow(query, workspace, activationCode).Scan(&updated)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Get the Slack workspaces of active accounts
func GetActiveWorkspaces(conn *sql.DB) ([]string, error) {
	rows, err := conn.Query(`SELECT slack_workspace FROM accounts WHERE account_expiry_date > NOW()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []string
	for rows.Next() {
		var ws string
		if err := rows.Scan(&ws); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

// Get the Slack workspaces and output channels of active accounts, so that created_at is older than the given digest range value
func GetActiveWorkspacesAndChannels(conn *sql.DB, dr models.DigestRange) ([]models.WorkspaceChannel, error) {
	query := `
        SELECT slack_workspace, slack_channel FROM accounts
        WHERE account_expiry_date > NOW()
		  AND slack_channel IS NOT NULL
          AND created_at < (CURRENT_DATE - CAST($1 as INTERVAL))
    `
	rows, err := conn.Query(query, string(dr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WorkspaceChannel
	for rows.Next() {
		var ws, ch string
		if err := rows.Scan(&ws, &ch); err != nil {
			return nil, err
		}
		result = append(result, models.WorkspaceChannel{Workspace: ws, Channel: ch})
	}
	return result, nil
}

// Get the prompt count for a workspace, checking if the provided prompt exists already
func GetPromptCountAndExists(conn *sql.DB, workspace, origin, category, prompt string) (count int, exists bool, err error) {
	err = conn.QueryRow(`
        SELECT COUNT(*) AS count,
               COALESCE(BOOL_OR(origin = $2 AND category = $3 AND prompt = $4), false) AS exists
        FROM prompts
        WHERE slack_workspace = $1
    `, workspace, origin, category, prompt).Scan(&count, &exists)
	return
}

// Get api_key and slack_channel for a given workspace
func GetAPIKeyAndChannel(conn *sql.DB, workspace string) (apiKey, channel string, err error) {
	var channelNS sql.NullString
	err = conn.QueryRow(`
        SELECT api_key, slack_channel
        FROM accounts
        WHERE slack_workspace = $1
        LIMIT 1
    `, workspace).Scan(&apiKey, &channelNS)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	if channelNS.Valid {
		channel = channelNS.String
	} else {
		channel = ""
	}
	return
}

// Update api_key for a given workspace
func UpdateApiKeyForWorkspace(conn *sql.DB, workspace, newApiKey string) error {
	_, err := conn.Exec(`
        UPDATE accounts
        SET api_key = $1
        WHERE slack_workspace = $2
    `, newApiKey, workspace)
	return err
}

// Update slack_channel for a given workspace
func UpdateSlackChannelForWorkspace(conn *sql.DB, workspace, channel string) error {
	_, err := conn.Exec(`
        UPDATE accounts
        SET slack_channel = $1
        WHERE slack_workspace = $2
    `, channel, workspace)
	return err
}

// Clear slack_channel for a given workspace
func ClearSlackChannelForWorkspace(conn *sql.DB, workspace string) error {
	_, err := conn.Exec(`
        UPDATE accounts
        SET slack_channel = NULL
        WHERE slack_workspace = $1
    `, workspace)
	return err
}
