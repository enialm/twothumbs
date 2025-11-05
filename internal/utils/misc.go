// File: internal/utils/misc.go

// This file contains miscellaneous utility functions.

package utils

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

func ConnectToDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return conn, nil
}

func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}

func PtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func TrimFeedbackFields(origin, category, prompt string) (string, string, string) {
	return strings.TrimSpace(origin), strings.TrimSpace(category), strings.TrimSpace(prompt)
}

func Spacer() map[string]any {
	return map[string]any{
		"type": "section",
		"text": map[string]any{
			"type": "plain_text",
			"text": " ",
		},
	}
}

func ParseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func IsDefaultDateRange(from, to time.Time) bool {
	defFrom := ParseDate(time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02"))
	defTo := ParseDate(time.Now().UTC().Format("2006-01-02"))
	return from.Equal(defFrom) && to.Equal(defTo)
}

func TimeToAgo(ts time.Time) string {
	duration := time.Since(ts)
	minutes := int(duration.Minutes())
	if minutes < 60 {
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	hours := int(duration.Hours())
	if hours < 24 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := hours / 24
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func FormatDelta(val float64) string {
	if val == 0 {
		return "-"
	}
	return fmt.Sprintf("%+.0f", math.Round(val))
}

func RoundFloat(val float64) int {
	return int(math.Round(val))
}

func FormatIssues(input string) string {
	paragraphs := strings.Split(input, "\n\n")
	for i, paragraph := range paragraphs {
		paragraphs[i] = fmt.Sprintf("%d. %s", i+1, strings.TrimSpace(paragraph))
	}
	return strings.Join(paragraphs, "\n\n\n")
}

func FormatDigestStats(
	prompt string,
	score int,
	scoreDelta string,
	nResponses int,
	nResponsesDelta string,
	nComments int,
	nCommentsDelta string,
) string {
	return fmt.Sprintf(
		"_%s_\n\n\n👍   %d%% (%s%%)    👋   %d (%s%%)    💬   %d (%s%%)",
		prompt,
		score,
		scoreDelta,
		nResponses,
		nResponsesDelta,
		nComments,
		nCommentsDelta,
	)
}

func FormatModalStats(
	category string,
	prompt string,
	score int,
	scoreDelta string,
	nResponses int,
	nResponsesDelta string,
	nComments int,
	nCommentsDelta string,
	printCategory bool,
) string {
	if printCategory {
		return fmt.Sprintf(
			"*%s*\n\n\n_%s_\n\n\n👍   %d%% (%s%%)\n\n👋   %d (%s%%)\n\n💬   %d (%s%%)",
			category,
			prompt,
			score,
			scoreDelta,
			nResponses,
			nResponsesDelta,
			nComments,
			nCommentsDelta,
		)
	} else {
		return fmt.Sprintf(
			"_%s_\n\n\n👍   %d%% (%s%%)\n\n👋   %d (%s%%)\n\n💬   %d (%s%%)",
			prompt,
			score,
			scoreDelta,
			nResponses,
			nResponsesDelta,
			nComments,
			nCommentsDelta,
		)
	}
}

func MonthLabel() string {
	now := time.Now().UTC()
	prevMonth := now.AddDate(0, -1, 0)
	return fmt.Sprintf("%s %d", prevMonth.Month().String(), prevMonth.Year())
}

func QuarterLabel() string {
	now := time.Now().UTC()
	month := int(now.Month())
	year := now.Year()

	currQ := (month-1)/3 + 1
	var prevQ, prevYear int
	if currQ == 1 {
		prevQ = 4
		prevYear = year - 1
	} else {
		prevQ = currQ - 1
		prevYear = year
	}
	return fmt.Sprintf("Q%d %d", prevQ, prevYear)
}

func IsMonday() bool {
	return time.Now().UTC().Weekday() == time.Monday
}

func IsFirstWeekdayOfMonth() bool {
	now := time.Now().UTC()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for first.Weekday() == time.Saturday || first.Weekday() == time.Sunday {
		first = first.AddDate(0, 0, 1)
	}
	return now.Year() == first.Year() && now.YearDay() == first.YearDay()
}

func IsSecondWeekdayOfQuarter() bool {
	now := time.Now().UTC()
	quarters := map[time.Month]bool{
		time.January: true, time.April: true, time.July: true, time.October: true,
	}
	if !quarters[now.Month()] {
		return false
	}
	// Find the first weekday of the month (quarter)
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for first.Weekday() == time.Saturday || first.Weekday() == time.Sunday {
		first = first.AddDate(0, 0, 1)
	}
	// Find the second weekday
	second := first.AddDate(0, 0, 1)
	for second.Weekday() == time.Saturday || second.Weekday() == time.Sunday {
		second = second.AddDate(0, 0, 1)
	}
	return now.Year() == second.Year() && now.YearDay() == second.YearDay()
}
