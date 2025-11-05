// File: internal/integrations/aiapi.go

// This file contains the logic to interact with the AI API for generating digests.

package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetAISummary(apiURL, apiKey, aiModel, aiPrompt, csvData string) (string, error) {
	payload := map[string]any{
		"model":        aiModel,
		"instructions": aiPrompt,
		"input":        csvData,
		"store":        false,
		"temperature":  0.0,
		"tool_choice":  "none",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API error %d: %s", resp.StatusCode, string(respBytes))
	}
	var result struct {
		Output []struct {
			Type    string `json:"type"`
			ID      string `json:"id"`
			Status  string `json:"status"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("unmarshal error: %w", err)
	}

	for _, msg := range result.Output {
		if msg.Type != "message" || msg.Status != "completed" || msg.Role != "assistant" {
			continue
		}
		for _, part := range msg.Content {
			if part.Type == "output_text" {
				return part.Text, nil
			}
		}
	}

	return "", fmt.Errorf("no output_text segment found")
}
