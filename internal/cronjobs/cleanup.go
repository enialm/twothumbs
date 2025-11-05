// File: internal/cronjobs/cleanup.go

// This file contains database cleanup functions to remove old data and update usage counts.

package cronjobs

import (
	"database/sql"
	"log"
)

func RunCleanup(conn *sql.DB) error {
	// Reset feedback_count values to zero
	if _, err := conn.Exec(`UPDATE accounts SET feedback_count = 0`); err != nil {
		log.Printf("Failed to reset feedback_count: %v", err)
		return err
	}

	// Delete feedback and summaries older than six months
	resf, err := conn.Exec(`DELETE FROM feedback WHERE DATE(created_at) < (CURRENT_DATE - INTERVAL '6 months')`)
	if err != nil {
		log.Printf("Failed to delete old feedback: %v", err)
		return err
	}
	n, _ := resf.RowsAffected()
	log.Printf("Deleted %d rows of old feedback data.", n)
	ress, err := conn.Exec(`DELETE FROM summaries WHERE DATE(summary_date) < (CURRENT_DATE - INTERVAL '6 months')`)
	if err != nil {
		log.Printf("Failed to delete old summaries: %v", err)
		return err
	}
	m, _ := ress.RowsAffected()
	log.Printf("Deleted %d rows of old cache data.", m)

	// Delete expired accounts and their data
	rows, err := conn.Query(`SELECT slack_workspace FROM accounts WHERE DATE(acccount_expiry_date) < (CURRENT_DATE - INTERVAL '1 month')`)
	if err != nil {
		log.Printf("Failed to get expired accounts: %v", err)
		return err
	}
	defer rows.Close()

	var workspaces []string
	for rows.Next() {
		var ws string
		if err := rows.Scan(&ws); err == nil && ws != "" {
			workspaces = append(workspaces, ws)
		}
	}

	for _, ws := range workspaces {
		if _, err := conn.Exec(`DELETE FROM prompts WHERE slack_workspace = $1`, ws); err != nil {
			log.Printf("Failed to delete prompts for workspace %s: %v", ws, err)
		}
		if _, err := conn.Exec(`DELETE FROM feedback WHERE slack_workspace = $1`, ws); err != nil {
			log.Printf("Failed to delete feedback for workspace %s: %v", ws, err)
		}
		if _, err := conn.Exec(`DELETE FROM summaries WHERE slack_workspace = $1`, ws); err != nil {
			log.Printf("Failed to delete summaries for workspace %s: %v", ws, err)
		}
		if _, err := conn.Exec(`DELETE FROM issues WHERE slack_workspace = $1`, ws); err != nil {
			log.Printf("Failed to delete issues for workspace %s: %v", ws, err)
		}
	}

	resc, err := conn.Exec(`DELETE FROM accounts WHERE DATE(account_expiry_date) < (CURRENT_DATE - INTERVAL '1 month')`)
	if err != nil {
		log.Printf("Failed to delete expired accounts: %v", err)
		return err
	}
	j, _ := resc.RowsAffected()
	log.Printf("Deleted %d expired accounts.", j)
	return nil
}
