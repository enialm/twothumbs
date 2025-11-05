// File: internal/db/queries/feedback.go

// This file shall contain queries related to feedback analysis.

package queries

import (
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"

	"twothumbs/internal/models"
	"twothumbs/internal/utils"
)

// Insert a new prompt into the database
func InsertPrompt(conn *sql.DB, workspace, origin, category, prompt string) (bool, error) {
	origin, category, prompt = utils.TrimFeedbackFields(origin, category, prompt)
	result, err := conn.Exec(`
        INSERT INTO prompts (
            slack_workspace, origin, category, prompt
        ) VALUES (
            $1, $2, $3, $4
        )
        ON CONFLICT (slack_workspace, origin, category, prompt) DO NOTHING
    `, workspace, origin, category, prompt)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// Delete a prompt and all related feedback from the database
func DeletePromptAndFeedback(conn *sql.DB, promptID string) error {
	// First, get the identifying fields for the prompt
	var workspace, origin, category, prompt string
	err := conn.QueryRow(`
        SELECT slack_workspace, origin, category, prompt
        FROM prompts
        WHERE id = $1
    `, promptID).Scan(&workspace, &origin, &category, &prompt)
	if err != nil {
		return err
	}

	// Delete all related feedback
	_, err = conn.Exec(`
        DELETE FROM feedback
        WHERE slack_workspace = $1
          AND origin = $2
          AND category = $3
          AND prompt = $4
    `, workspace, origin, category, prompt)
	if err != nil {
		return err
	}

	// Delete all related summaries
	_, err = conn.Exec(`
        DELETE FROM summaries
        WHERE slack_workspace = $1
          AND origin = $2
          AND category = $3
          AND prompt = $4
    `, workspace, origin, category, prompt)
	if err != nil {
		return err
	}

	// Delete the prompt itself
	_, err = conn.Exec(`
        DELETE FROM prompts
        WHERE id = $1
    `, promptID)
	return err
}

// Get a prompt by its ID
func GetPrompt(conn *sql.DB, promptID string) (models.Prompt, error) {
	var p models.Prompt
	err := conn.QueryRow(`
        SELECT id, slack_workspace, origin, category, prompt
        FROM prompts
        WHERE id = $1
    `, promptID).Scan(
		&p.ID,
		&p.SlackWorkspace,
		&p.Origin,
		&p.Category,
		&p.Prompt,
	)
	return p, err
}

// Get all prompts for a given workspace, ordered by origin, category, and prompt
func GetPrompts(conn *sql.DB, workspace string) ([]models.Prompt, error) {
	rows, err := conn.Query(`
        SELECT id, slack_workspace, origin, category, prompt
        FROM prompts
        WHERE slack_workspace = $1
        ORDER BY origin, category, prompt
    `, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var p models.Prompt
		if err := rows.Scan(
			&p.ID,
			&p.SlackWorkspace,
			&p.Origin,
			&p.Category,
			&p.Prompt,
		); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}
	return prompts, nil
}

// Insert a new feedback record into the database and increment the feedback count for the account.
func InsertFeedback(conn *sql.DB, fb *models.Feedback) error {
	fb.Origin, fb.Category, fb.Prompt = utils.TrimFeedbackFields(fb.Origin, fb.Category, fb.Prompt)
	if fb.Comment != nil {
		trimmed := strings.TrimSpace(*fb.Comment)
		if trimmed == "" {
			fb.Comment = nil
		} else {
			fb.Comment = &trimmed
		}
	}
	_, err := conn.Exec(`
        INSERT INTO feedback (
            slack_workspace, prompt, thumb_up, comment, origin, category, in_production, user_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        )
    `,
		fb.SlackWorkspace,
		fb.Prompt,
		fb.ThumbUp,
		fb.Comment,
		fb.Origin,
		fb.Category,
		fb.InProduction,
		strings.TrimSpace(fb.UserID),
	)
	if err != nil {
		return err
	}

	// Increment feedback_count for the account
	_, err = conn.Exec(`
        UPDATE accounts
        SET feedback_count = COALESCE(feedback_count, 0) + 1
        WHERE slack_workspace = $1
    `, fb.SlackWorkspace)
	return err
}

// Get distinct origins, categories, prompts, and feedback count for a workspace and time range, with filters
func GetHomeTabData(
	conn *sql.DB,
	workspace string,
	from, to time.Time,
	origin, category, prompt, thumb string,
) ([]string, []string, []string, int, int, error) {
	query := `
    WITH
    origins AS (
        SELECT DISTINCT origin FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND created_at >= $2
          AND created_at < $3
          AND ($5 = '' OR category = $5)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
    ),
    categories AS (
        SELECT DISTINCT category FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
    ),
    prompts AS (
        SELECT DISTINCT prompt FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($5 = '' OR category = $5)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
    ),
    counts AS (
        SELECT COUNT(*) AS feedback_count FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($5 = '' OR category = $5)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
    ),
    comments AS (
        SELECT COUNT(*) AS comment_count FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND comment IS NOT NULL
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($5 = '' OR category = $5)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
    )
    SELECT
        ARRAY(SELECT origin FROM origins),
        ARRAY(SELECT category FROM categories),
        ARRAY(SELECT prompt FROM prompts),
        (SELECT feedback_count FROM counts),
        (SELECT comment_count FROM comments)
    `

	var outOrigins, outCategories, outPrompts []string
	var feedbackCount, commentCount int
	row := conn.QueryRow(query, workspace, from, to, origin, category, prompt, thumb)
	err := row.Scan(pq.Array(&outOrigins), pq.Array(&outCategories), pq.Array(&outPrompts), &feedbackCount, &commentCount)
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}
	return outOrigins, outCategories, outPrompts, feedbackCount, commentCount, nil
}

// Get distinct feedback groups for a given workspace and time range
func GetFeedbackGroups(conn *sql.DB, workspace string, dr models.DigestRange) ([]models.FeedbackGroup, error) {
	var query string
	switch dr {
	case models.Quarterly:
		query = `
                SELECT DISTINCT origin
                FROM feedback
                WHERE slack_workspace = $1
                  AND in_production = TRUE
                  AND DATE(created_at) >= DATE_TRUNC('month', CURRENT_DATE - CAST($2 AS INTERVAL))
                  AND DATE(created_at) < DATE_TRUNC('month', CURRENT_DATE)
            `
	case models.Monthly:
		query = `
                SELECT DISTINCT origin, category
                FROM feedback
                WHERE slack_workspace = $1
                  AND in_production = TRUE
                  AND DATE(created_at) >= DATE_TRUNC('month', CURRENT_DATE - CAST($2 AS INTERVAL))
                  AND DATE(created_at) < DATE_TRUNC('month', CURRENT_DATE)
            `
	case models.Last7d: // Last 7 days
		query = `
                SELECT DISTINCT origin, category, prompt
                FROM feedback
                WHERE slack_workspace = $1
                  AND in_production = TRUE
                  AND created_at >= (CURRENT_DATE - CAST($2 AS INTERVAL))
            `
	case models.Last30d: // Last 30 days
		query = `
                SELECT DISTINCT origin, category, prompt
                FROM feedback
                WHERE slack_workspace = $1
                  AND in_production = TRUE
                  AND created_at >= (CURRENT_DATE - CAST($2 AS INTERVAL))
            `
	default: // Weekly
		query = `
                SELECT DISTINCT origin, category, prompt
                FROM feedback
                WHERE slack_workspace = $1
                  AND in_production = TRUE
                  AND DATE(created_at) >= (CURRENT_DATE - CAST($2 AS INTERVAL))
                  AND DATE(created_at) < CURRENT_DATE
            `
	}
	rows, err := conn.Query(query, workspace, string(dr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.FeedbackGroup
	for rows.Next() {
		var g models.FeedbackGroup
		switch dr {
		case models.Quarterly:
			if err := rows.Scan(&g.Origin); err != nil {
				return nil, err
			}
		case models.Monthly:
			if err := rows.Scan(&g.Origin, &g.Category); err != nil {
				return nil, err
			}
		default: // Weekly
			if err := rows.Scan(&g.Origin, &g.Category, &g.Prompt); err != nil {
				return nil, err
			}
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// Get distinct feedback groups for a workspace, date range, and optional filters
func GetFeedbackGroupsWithFilters(
	conn *sql.DB,
	workspace string,
	from, to time.Time,
	origin, category, prompt, thumb string,
) ([]models.FeedbackGroup, error) {
	query := `
        SELECT DISTINCT origin, category, prompt
        FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($5 = '' OR category = $5)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
    `
	args := []any{workspace, from, to, origin, category, prompt, thumb}

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.FeedbackGroup
	for rows.Next() {
		var g models.FeedbackGroup
		if err := rows.Scan(&g.Origin, &g.Category, &g.Prompt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// Fetch summaries for a given workspace and time range
func GetSummariesByWorkspace(conn *sql.DB, workspace string, dr models.DigestRange) ([]models.SummaryRow, error) {
	var query string
	switch dr {
	case models.Monthly:
		query = `
            SELECT summary_date, origin, category, prompt, n_comments, summary
            FROM summaries
            WHERE slack_workspace = $1
              AND summary_date >= DATE_TRUNC('month', CURRENT_DATE - CAST($2 AS INTERVAL))
              AND summary_date < DATE_TRUNC('month', CURRENT_DATE)
        `
	case models.Quarterly:
		query = `
            SELECT summary_date, origin, category, prompt, n_comments, summary
            FROM summaries
            WHERE slack_workspace = $1
              AND summary_date >= DATE_TRUNC('month', CURRENT_DATE - CAST($2 AS INTERVAL))
              AND summary_date < DATE_TRUNC('month', CURRENT_DATE)
        `
	default: // Daily or Weekly
		query = `
            SELECT summary_date, origin, category, prompt, n_comments, summary
            FROM summaries
            WHERE slack_workspace = $1
              AND summary_date >= (CURRENT_DATE - CAST($2 AS INTERVAL))
              AND summary_date < CURRENT_DATE
        `
	}

	var rows *sql.Rows
	var err error
	rows, err = conn.Query(query, workspace, string(dr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.SummaryRow
	for rows.Next() {
		var s models.SummaryRow
		if err := rows.Scan(&s.SummaryDate, &s.Origin, &s.Category, &s.Prompt, &s.NComments, &s.Summary); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// Calculate thumbs up %, feedback count, and comment count for the current and previous period
// by origin, category, and prompt
func GetWeeklyFeedbackStats(conn *sql.DB, workspace string, origin string, category string, prompt string) (*models.FeedbackStats, error) {
	query := `
    WITH
    current_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND DATE(created_at) >= (CURRENT_DATE - INTERVAL '7 days')
              AND DATE(created_at) < CURRENT_DATE
            ORDER BY user_id, created_at DESC
        ) latest
    ),
    prev_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND DATE(created_at) >= (CURRENT_DATE - INTERVAL '14 days')
              AND DATE(created_at) < (CURRENT_DATE - INTERVAL '7 days')
            ORDER BY user_id, created_at DESC
        ) latest
    )
    SELECT
        CASE WHEN c.num_feedback > 0
            THEN 100.0 * c.thumbs_up / c.num_feedback
            ELSE 0 END AS thumbs_up_pct,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * p.thumbs_up / p.num_feedback
            ELSE 0 END AS prev_thumbs_up_pct,
        c.num_feedback AS n_feedback,
        p.num_feedback AS prev_n_feedback,
		CASE WHEN p.num_feedback > 0
        	THEN 100.0 * (c.num_feedback - p.num_feedback) / p.num_feedback
        ELSE 0 END AS prev_n_feedback_dl,
        c.num_comments,
		CASE WHEN p.num_comments > 0
			THEN 100.0 * (c.num_comments - p.num_comments) / p.num_comments
		ELSE 0 END AS prev_comments_dl
    FROM
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM current_period
        ) c,
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM prev_period
        ) p;
    `

	var stats models.FeedbackStats
	err := conn.QueryRow(query, workspace, origin, category, prompt).Scan(
		&stats.ThumbsUpPct,
		&stats.PrevThumbsUpPct,
		&stats.NFeedback,
		&stats.PrevNFeedback,
		&stats.PrevNFeedbackDl,
		&stats.NComments,
		&stats.PrevCommentsDl,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Calculate thumbs up %, feedback count, and comment count for the current and previous period
// by origin and category
func GetMonthlyFeedbackStats(conn *sql.DB, workspace string, origin string, category string) (*models.FeedbackStats, error) {
	query := `
    WITH
    current_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND DATE(created_at) >= (DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '1 month')
              AND DATE(created_at) < DATE_TRUNC('month', CURRENT_DATE)
            ORDER BY user_id, created_at DESC
        ) latest
    ),
    prev_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND DATE(created_at) >= (DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '2 months')
              AND DATE(created_at) < (DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '1 month')
            ORDER BY user_id, created_at DESC
        ) latest
    )
    SELECT
        CASE WHEN c.num_feedback > 0
            THEN 100.0 * c.thumbs_up / c.num_feedback
            ELSE 0 END AS thumbs_up_pct,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * p.thumbs_up / p.num_feedback
            ELSE 0 END AS prev_thumbs_up_pct,
        c.num_feedback AS n_feedback,
        p.num_feedback AS prev_n_feedback,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * (c.num_feedback - p.num_feedback) / p.num_feedback
        ELSE 0 END AS prev_n_feedback_dl,
        c.num_comments,
        CASE WHEN p.num_comments > 0
            THEN 100.0 * (c.num_comments - p.num_comments) / p.num_comments
        ELSE 0 END AS prev_comments_dl
    FROM
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM current_period
        ) c,
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM prev_period
        ) p;
    `

	var stats models.FeedbackStats
	err := conn.QueryRow(query, workspace, origin, category).Scan(
		&stats.ThumbsUpPct,
		&stats.PrevThumbsUpPct,
		&stats.NFeedback,
		&stats.PrevNFeedback,
		&stats.PrevNFeedbackDl,
		&stats.NComments,
		&stats.PrevCommentsDl,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Get feedback stats for plotting the monthly plots
func GetMonthlyDigestPlotStats(conn *sql.DB, workspace, origin, category string) ([]models.PlotStats, error) {
	query := `
        SELECT
            DATE_TRUNC('month', created_at) AS month,
            CASE WHEN COUNT(thumb_up) > 0
                THEN 100.0 * COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) / COUNT(thumb_up)
                ELSE 0 END AS thumbs_up_pct,
            COUNT(thumb_up) AS n_feedback,
            COUNT(comment) FILTER (WHERE comment IS NOT NULL) AS n_comments
        FROM feedback
        WHERE slack_workspace = $1
          AND origin = $2
          AND category = $3
          AND in_production = TRUE
          AND DATE(created_at) >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '6 months')
          AND DATE(created_at) < DATE_TRUNC('month', CURRENT_DATE)
        GROUP BY month
        ORDER BY month
    `
	rows, err := conn.Query(query, workspace, origin, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PlotStats
	for rows.Next() {
		var s models.PlotStats
		if err := rows.Scan(&s.Month, &s.ThumbsUpPct, &s.NFeedback, &s.NComments); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// Get feedback stats for plotting the quarterly plots
func GetQuarterlyDigestPlotStats(conn *sql.DB, workspace string, origin string) ([]models.PlotStats, error) {
	query := `
        SELECT
            DATE_TRUNC('month', created_at) AS month,
            CASE WHEN COUNT(thumb_up) > 0
                THEN 100.0 * COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) / COUNT(thumb_up)
                ELSE 0 END AS thumbs_up_pct,
            COUNT(thumb_up) AS n_feedback,
            COUNT(comment) FILTER (WHERE comment IS NOT NULL) AS n_comments
        FROM feedback
        WHERE slack_workspace = $1
          AND origin = $2
          AND in_production = TRUE
          AND DATE(created_at) >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '6 months')
          AND DATE(created_at) < DATE_TRUNC('month', CURRENT_DATE)
        GROUP BY month
        ORDER BY month
    `
	rows, err := conn.Query(query, workspace, origin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PlotStats
	for rows.Next() {
		var s models.PlotStats
		if err := rows.Scan(&s.Month, &s.ThumbsUpPct, &s.NFeedback, &s.NComments); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// Get the n latest comments for a workspace
func GetLatestComments(conn *sql.DB, workspace string, n int) ([]models.ExploreCommentsResult, error) {
	query := `
        SELECT origin, category, prompt, comment, created_at
        FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND comment IS NOT NULL
        ORDER BY created_at DESC
        LIMIT $2
    `
	rows, err := conn.Query(query, workspace, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ExploreCommentsResult
	for rows.Next() {
		var r models.ExploreCommentsResult
		if err := rows.Scan(&r.Origin, &r.Category, &r.Prompt, &r.Comment, &r.Timestamp); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Get comments with filters
func GetCommentsWithFilters(
	conn *sql.DB,
	workspace string,
	from, to time.Time,
	origin, category, prompt, thumb string,
	limit int,
) ([]models.ExploreCommentsResult, error) {
	query := `
        SELECT origin, category, prompt, comment, created_at
        FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND comment IS NOT NULL
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($5 = '' OR category = $5)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
        ORDER BY created_at DESC
        LIMIT $8
    `
	rows, err := conn.Query(query, workspace, from, to, origin, category, prompt, thumb, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ExploreCommentsResult
	for rows.Next() {
		var r models.ExploreCommentsResult
		if err := rows.Scan(&r.Origin, &r.Category, &r.Prompt, &r.Comment, &r.Timestamp); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Get top issues reported by accounts for a workspace
func GetTopIssuesByWorkspace(conn *sql.DB, workspace string) ([]models.IssueReport, error) {
	rows, err := conn.Query(`
        SELECT origin, report
        FROM issues
        WHERE slack_workspace = $1
        ORDER BY origin
    `, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []models.IssueReport
	for rows.Next() {
		var issue models.IssueReport
		if err := rows.Scan(&issue.Origin, &issue.Report); err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

// Calculate thumbs up %, feedback count, and comment count for the current and previous period by origin, category, and prompt
func GetLast7DayStats(conn *sql.DB, workspace string, origin string, category string, prompt string) (*models.FeedbackStats, error) {
	query := `
    WITH
    current_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND created_at >= (CURRENT_DATE - INTERVAL '7 days')
            ORDER BY user_id, created_at DESC
        ) latest
    ),
    prev_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND created_at >= (CURRENT_DATE - INTERVAL '14 days')
              AND DATE(created_at) <= (CURRENT_DATE - INTERVAL '7 days')
            ORDER BY user_id, created_at DESC
        ) latest
    )
    SELECT
        CASE WHEN c.num_feedback > 0
            THEN 100.0 * c.thumbs_up / c.num_feedback
            ELSE 0 END AS thumbs_up_pct,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * p.thumbs_up / p.num_feedback
            ELSE 0 END AS prev_thumbs_up_pct,
        c.num_feedback AS n_feedback,
        p.num_feedback AS prev_n_feedback,
		CASE WHEN p.num_feedback > 0
        	THEN 100.0 * (c.num_feedback - p.num_feedback) / p.num_feedback
        ELSE 0 END AS prev_n_feedback_dl,
        c.num_comments,
		CASE WHEN p.num_comments > 0
			THEN 100.0 * (c.num_comments - p.num_comments) / p.num_comments
		ELSE 0 END AS prev_comments_dl
    FROM
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM current_period
        ) c,
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM prev_period
        ) p;
    `

	var stats models.FeedbackStats
	err := conn.QueryRow(query, workspace, origin, category, prompt).Scan(
		&stats.ThumbsUpPct,
		&stats.PrevThumbsUpPct,
		&stats.NFeedback,
		&stats.PrevNFeedback,
		&stats.PrevNFeedbackDl,
		&stats.NComments,
		&stats.PrevCommentsDl,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Calculate thumbs up %, feedback count, and comment count for the current and previous period by origin, category, and prompt
func GetLast30DayStats(conn *sql.DB, workspace string, origin string, category string, prompt string) (*models.FeedbackStats, error) {
	query := `
    WITH
    current_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND created_at >= (CURRENT_DATE - INTERVAL '30 days')
            ORDER BY user_id, created_at DESC
        ) latest
    ),
    prev_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND created_at >= (CURRENT_DATE - INTERVAL '60 days')
              AND DATE(created_at) <= (CURRENT_DATE - INTERVAL '30 days')
            ORDER BY user_id, created_at DESC
        ) latest
    )
    SELECT
        CASE WHEN c.num_feedback > 0
            THEN 100.0 * c.thumbs_up / c.num_feedback
            ELSE 0 END AS thumbs_up_pct,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * p.thumbs_up / p.num_feedback
            ELSE 0 END AS prev_thumbs_up_pct,
        c.num_feedback AS n_feedback,
        p.num_feedback AS prev_n_feedback,
		CASE WHEN p.num_feedback > 0
        	THEN 100.0 * (c.num_feedback - p.num_feedback) / p.num_feedback
        ELSE 0 END AS prev_n_feedback_dl,
        c.num_comments,
		CASE WHEN p.num_comments > 0
			THEN 100.0 * (c.num_comments - p.num_comments) / p.num_comments
		ELSE 0 END AS prev_comments_dl
    FROM
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM current_period
        ) c,
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM prev_period
        ) p;
    `

	var stats models.FeedbackStats
	err := conn.QueryRow(query, workspace, origin, category, prompt).Scan(
		&stats.ThumbsUpPct,
		&stats.PrevThumbsUpPct,
		&stats.NFeedback,
		&stats.PrevNFeedback,
		&stats.PrevNFeedbackDl,
		&stats.NComments,
		&stats.PrevCommentsDl,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Calculate thumbs up %, feedback count, and comment count for the current and previous period for a given workspace, date range, and filters (origin, category, prompt, thumb)
func GetFilteredStats(
	conn *sql.DB,
	workspace string,
	origin, category, prompt string,
	from, to time.Time,
	prevFrom, prevTo time.Time,
	thumb string,
) (*models.FeedbackStats, error) {
	query := `
    WITH
    current_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND created_at >= $5
              AND created_at < $6
              AND (
                $9 = '' OR
                ($9 = 'Up' AND thumb_up = TRUE) OR
                ($9 = 'Down' AND thumb_up = FALSE)
              )
            ORDER BY user_id, created_at DESC
        ) latest
    ),
    prev_period AS (
        SELECT *
        FROM (
            SELECT DISTINCT ON (user_id)
                user_id,
                thumb_up,
                comment,
                created_at
            FROM feedback
            WHERE slack_workspace = $1
              AND in_production = TRUE
              AND origin = $2
              AND category = $3
              AND prompt = $4
              AND created_at >= $7
              AND created_at < $8
              AND (
                $9 = '' OR
                ($9 = 'Up' AND thumb_up = TRUE) OR
                ($9 = 'Down' AND thumb_up = FALSE)
              )
            ORDER BY user_id, created_at DESC
        ) latest
    )
    SELECT
        CASE WHEN c.num_feedback > 0
            THEN 100.0 * c.thumbs_up / c.num_feedback
            ELSE 0 END AS thumbs_up_pct,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * p.thumbs_up / p.num_feedback
            ELSE 0 END AS prev_thumbs_up_pct,
        c.num_feedback AS n_feedback,
        p.num_feedback AS prev_n_feedback,
        CASE WHEN p.num_feedback > 0
            THEN 100.0 * (c.num_feedback - p.num_feedback) / p.num_feedback
        ELSE 0 END AS prev_n_feedback_dl,
        c.num_comments,
        CASE WHEN p.num_comments > 0
            THEN 100.0 * (c.num_comments - p.num_comments) / p.num_comments
        ELSE 0 END AS prev_comments_dl
    FROM
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM current_period
        ) c,
        (SELECT
            COUNT(thumb_up) FILTER (WHERE thumb_up IS TRUE) AS thumbs_up,
            COUNT(thumb_up) AS num_feedback,
            COUNT(comment) AS num_comments
         FROM prev_period
        ) p;
    `

	var stats models.FeedbackStats
	err := conn.QueryRow(
		query,
		workspace, origin, category, prompt,
		from, to,
		prevFrom, prevTo,
		thumb,
	).Scan(
		&stats.ThumbsUpPct,
		&stats.PrevThumbsUpPct,
		&stats.NFeedback,
		&stats.PrevNFeedback,
		&stats.PrevNFeedbackDl,
		&stats.NComments,
		&stats.PrevCommentsDl,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Get raw feedback data for a workspace and time range, with optional filters
func GetRawFeedbackData(
	conn *sql.DB,
	workspace string,
	from, to time.Time,
	origin, category, prompt, thumb string,
) ([]models.RawData, error) {
	query := `
        SELECT
            created_at,
            origin,
            category,
            prompt,
            thumb_up,
            comment,
            user_id
        FROM feedback
        WHERE slack_workspace = $1
          AND in_production = TRUE
          AND created_at >= $2
          AND created_at < $3
          AND ($4 = '' OR origin = $4)
          AND ($5 = '' OR category = $5)
          AND ($6 = '' OR prompt = $6)
          AND (
            $7 = '' OR
            ($7 = 'Up' AND thumb_up = TRUE) OR
            ($7 = 'Down' AND thumb_up = FALSE)
          )
        ORDER BY created_at DESC
    `
	rows, err := conn.Query(query, workspace, from, to, origin, category, prompt, thumb)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.RawData
	for rows.Next() {
		var fb models.RawData
		err := rows.Scan(
			&fb.CreatedAt,
			&fb.Origin,
			&fb.Category,
			&fb.Prompt,
			&fb.ThumbUp,
			&fb.Comment,
			&fb.UserID,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, fb)
	}
	return results, nil
}

// Get test feedback data for a workspace
func GetTestFeedbackData(
	conn *sql.DB,
	workspace string,
) ([]models.RawData, error) {
	query := `
        SELECT
            created_at,
            origin,
            category,
            prompt,
            thumb_up,
            comment,
            user_id
        FROM feedback
        WHERE slack_workspace = $1
          AND in_production = FALSE
        ORDER BY created_at DESC
        LIMIT 10
    `
	rows, err := conn.Query(query, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.RawData
	for rows.Next() {
		var fb models.RawData
		err := rows.Scan(
			&fb.CreatedAt,
			&fb.Origin,
			&fb.Category,
			&fb.Prompt,
			&fb.ThumbUp,
			&fb.Comment,
			&fb.UserID,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, fb)
	}
	return results, nil
}
