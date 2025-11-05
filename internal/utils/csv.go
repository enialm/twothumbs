// File: internal/utils/csv.go

// This file contains utility functions for CSV formatting of summaries and feedback.

package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"twothumbs/internal/models"
)

// CSV formatting for digests and top issues
func SummariesToCSV(summaries []models.SummaryRow, withDates bool) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if withDates {
		w.Write([]string{"summary_date", "category", "prompt", "n_comments", "summary"})
		for _, s := range summaries {
			w.Write([]string{
				s.SummaryDate.Format("2006-01-02"),
				s.Category,
				s.Prompt,
				fmt.Sprintf("%d", s.NComments),
				s.Summary,
			})
		}
	} else {
		w.Write([]string{"category", "prompt", "n_comments", "summary"})
		for _, s := range summaries {
			w.Write([]string{
				s.Category,
				s.Prompt,
				fmt.Sprintf("%d", s.NComments),
				s.Summary,
			})
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}

// CSV formatting for comment caching
func FeedbackToCSV(feedbacks []*models.Feedback) (string, error) {
	if len(feedbacks) == 0 {
		return "", nil
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write the prompt
	w.Write([]string{"prompt"})
	w.Write([]string{feedbacks[0].Prompt})
	w.Write([]string{}) // An empty line for separation

	// Write the feedback block
	w.Write([]string{"comment", "user_id"})
	for _, f := range feedbacks {
		w.Write([]string{
			PtrToString(f.Comment),
			f.UserID,
		})
	}
	w.Flush()
	return buf.String(), w.Error()
}

func RawDataToCSV(data []models.RawData) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header
	w.Write([]string{
		"created_at", "origin", "category", "prompt", "thumb_up", "comment", "user_id",
	})

	// Write data rows
	for _, r := range data {
		comment := ""
		if r.Comment != nil {
			comment = *r.Comment
		}
		w.Write([]string{
			r.CreatedAt.Format(time.RFC3339),
			r.Origin,
			r.Category,
			r.Prompt,
			fmt.Sprintf("%t", r.ThumbUp),
			comment,
			r.UserID,
		})
	}
	w.Flush()
	return buf.String(), w.Error()
}
