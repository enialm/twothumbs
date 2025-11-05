// File: internal/db/queries/cronjobs.go

// This file contains the database queries related to cronjobs.

package queries

import (
	"database/sql"

	"twothumbs/internal/models"
)

// Get comments for caching
func GetCommentsForCaching(conn *sql.DB, workspace string) ([]*models.Feedback, error) {
	rows, err := conn.Query(`
        SELECT prompt, thumb_up, comment, origin, category, user_id
        FROM feedback
        WHERE slack_workspace = $1
            AND in_production = true
            AND comment IS NOT NULL
            AND DATE(created_at) = CURRENT_DATE - INTERVAL '1 day'
    `, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []*models.Feedback
	for rows.Next() {
		var f models.Feedback
		if err := rows.Scan(&f.Prompt, &f.ThumbUp, &f.Comment, &f.Origin, &f.Category, &f.UserID); err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, &f)
	}
	return feedbacks, nil
}

// Insert a summary into the database
func InsertSummary(conn *sql.DB, workspace, prompt, origin, category string, nComments int, summary string) error {
	_, err := conn.Exec(`
        INSERT INTO summaries (summary_date, slack_workspace, origin, category, prompt, n_comments, summary)
        VALUES (CURRENT_DATE - INTERVAL '1 day', $1, $2, $3, $4, $5, $6)
    `, workspace, origin, category, prompt, nComments, summary)
	return err
}

// Fetch distinct origins for a workspace
func GetDistinctOrigins(conn *sql.DB, workspace string) ([]string, error) {
	rows, err := conn.Query(`
        SELECT DISTINCT origin
        FROM summaries
        WHERE slack_workspace = $1
    `, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var origins []string
	for rows.Next() {
		var origin string
		if err := rows.Scan(&origin); err != nil {
			return nil, err
		}
		origins = append(origins, origin)
	}
	return origins, nil
}

// Fetch summaries from the last 30 days for a given workspace
func GetLatestSummaries(conn *sql.DB, workspace string) ([]*models.SummaryRow, error) {
	rows, err := conn.Query(`
        SELECT summary_date, origin, category, prompt, n_comments, summary
        FROM summaries
        WHERE slack_workspace = $1
        AND summary_date >= CURRENT_DATE - INTERVAL '30 day'
        ORDER BY summary_date DESC
    `, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.SummaryRow
	for rows.Next() {
		var f models.SummaryRow
		if err := rows.Scan(&f.SummaryDate, &f.Origin, &f.Category, &f.Prompt, &f.NComments, &f.Summary); err != nil {
			return nil, err
		}
		summaries = append(summaries, &f)
	}
	return summaries, nil
}

// Insert an issue report into the database
func InsertIssueReport(conn *sql.DB, workspace, origin, issueReport string) error {
	_, err := conn.Exec(`
        INSERT INTO issues (slack_workspace, origin, report)
        VALUES ($1, $2, $3)
        ON CONFLICT (slack_workspace, origin)
        DO UPDATE SET report = EXCLUDED.report
    `, workspace, origin, issueReport)
	return err
}

// Truncate the issues table
func TruncateIssuesTable(conn *sql.DB) error {
	_, err := conn.Exec(`TRUNCATE TABLE issues`)
	return err
}
