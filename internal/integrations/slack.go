// File: internal/integrations/slack.go

// This file contains logic to interact with Slack.

package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"twothumbs/internal/models"
)

const (
	maxBlocks = 50
	maxChars  = 13200 // Slack bug: limit of 13200 chars for rich_text messages
)

// Send a Block Kit message to a Slack channel using a bot token
func SendBlockKitMessage(botToken, channel string, blocks []map[string]any) error {
	chunks := chunkBlocks(blocks, channel)

	for i, chunk := range chunks {
		if err := sendSingleMessage(botToken, channel, chunk); err != nil {
			return fmt.Errorf("failed to send message part %d/%d: %w", i+1, len(chunks), err)
		}
	}

	return nil
}

// Send a single Block Kit message to Slack
func sendSingleMessage(botToken, channel string, blocks []map[string]any) error {
	payload := map[string]any{
		"channel": channel,
		"blocks":  blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("slack API returned status: %s", resp.Status)
	}

	// Decode the response body to check for errors
	var respBody struct {
		OK               bool   `json:"ok"`
		Error            string `json:"error,omitempty"`
		ResponseMetadata struct {
			Messages []string `json:"messages"`
		} `json:"response_metadata,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to decode Slack response: %w", err)
	}
	if !respBody.OK {
		metaMsg := ""
		if len(respBody.ResponseMetadata.Messages) > 0 {
			metaMsg = fmt.Sprintf(" Details: %v", respBody.ResponseMetadata.Messages)
		}
		return fmt.Errorf("slack API error: %s.%s", respBody.Error, metaMsg)
	}

	return nil
}

// Make sure blocks are split into chunks that fit Slack's limits
func chunkBlocks(blocks []map[string]any, channel string) [][]map[string]any {
	var chunks [][]map[string]any

	var split func([]map[string]any)
	split = func(bs []map[string]any) {
		if len(bs) == 0 {
			return
		}
		payload := map[string]any{
			"channel": channel,
			"blocks":  bs,
		}
		body, _ := json.Marshal(payload)
		if len(bs) <= maxBlocks && len(body) <= maxChars {
			chunks = append(chunks, bs)
			return
		}
		if len(bs) == 1 {
			chunks = append(chunks, bs)
			return
		}
		mid := len(bs) / 2
		split(bs[:mid])
		split(bs[mid:])
	}

	split(blocks)
	return chunks
}

// Upload a file to Slack
func UploadFileToSlack(botToken, channel, filePath, title, initialComment string) (string, error) {
	// Get the file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()

	// Get the upload URL
	req, err := http.NewRequest("GET", fmt.Sprintf("https://slack.com/api/files.getUploadURLExternal?filename=%s&length=%d", url.QueryEscape(fileName), fileSize), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create getUploadURL request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var uploadURLResp models.SlackUploadURLResponse
	if err := json.Unmarshal(respBody, &uploadURLResp); err != nil {
		return "", fmt.Errorf("failed to decode getUploadURL response: %w", err)
	}
	if !uploadURLResp.OK {
		metaMsg := ""
		if len(uploadURLResp.ResponseMetadata.Messages) > 0 {
			metaMsg = fmt.Sprintf(" Details: %v", uploadURLResp.ResponseMetadata.Messages)
		}
		return "", fmt.Errorf("slack getUploadURL error: %s.%s", uploadURLResp.Error, metaMsg)
	}

	uploadURL := uploadURLResp.UploadURL
	fileID := uploadURLResp.FileID

	// Upload the file to the provided URL
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var fileUploadBuffer bytes.Buffer
	writer := multipart.NewWriter(&fileUploadBuffer)
	formFile, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(formFile, file); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}
	writer.Close()

	uploadReq, err := http.NewRequest("POST", uploadURL, &fileUploadBuffer)
	if err != nil {
		return "", fmt.Errorf("failed to create file upload request: %w", err)
	}
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Slack: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != 200 {
		return "", fmt.Errorf("file upload to Slack failed with status: %d", uploadResp.StatusCode)
	}

	// Complete the file upload
	completePayload := map[string]any{
		"files": []map[string]string{
			{
				"id":    fileID,
				"title": title,
			},
		},
	}

	// If an initial comment is provided, send immediately
	if initialComment != "" {
		completePayload["channel_id"] = channel
		completePayload["initial_comment"] = initialComment
	}

	completePayloadBytes, _ := json.Marshal(completePayload)

	completeReq, err := http.NewRequest("POST", "https://slack.com/api/files.completeUploadExternal", bytes.NewReader(completePayloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create completeUpload request: %w", err)
	}
	completeReq.Header.Set("Authorization", "Bearer "+botToken)
	completeReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	completeResp, err := http.DefaultClient.Do(completeReq)
	if err != nil {
		return "", fmt.Errorf("failed to complete file upload: %w", err)
	}
	defer completeResp.Body.Close()

	completeRespBody, _ := io.ReadAll(completeResp.Body)

	var completeUploadResp models.SlackCompleteUploadResponse
	if err := json.Unmarshal(completeRespBody, &completeUploadResp); err != nil {
		return "", fmt.Errorf("failed to decode completeUpload response: %w", err)
	}
	if !completeUploadResp.OK {
		metaMsg := ""
		if len(completeUploadResp.ResponseMetadata.Messages) > 0 {
			metaMsg = fmt.Sprintf(" Details: %v", completeUploadResp.ResponseMetadata.Messages)
		}
		return "", fmt.Errorf("slack completeUpload error: %s.%s", completeUploadResp.Error, metaMsg)
	}

	if len(completeUploadResp.Files) == 0 {
		return "", fmt.Errorf("no files returned from upload")
	}

	responseFiles := completeUploadResp.Files[0]

	// Return the private URL for attaching to messages later
	return responseFiles.URLPrivate, nil
}

// Push a new Slack modal using views.push
func PushSlackView(triggerID string, modal map[string]any, botToken string) error {
	// Construct the payload for the views.push API
	payload := map[string]any{
		"trigger_id": triggerID,
		"view":       modal,
	}

	// Make the HTTP request to Slack's API
	slackURL := "https://slack.com/api/views.push"
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Add the Authorization header with the bot token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response from Slack
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var respData map[string]any
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	// Check if the API call was successful
	if ok, okType := respData["ok"].(bool); !okType || !ok {
		metaMsg := ""
		if meta, ok := respData["response_metadata"].(map[string]any); ok {
			if msgs, ok := meta["messages"].([]any); ok && len(msgs) > 0 {
				metaMsg = fmt.Sprintf(" Details: %v", msgs)
			}
		}
		return fmt.Errorf("slack API error: %v.%s", respData["error"], metaMsg)
	}

	return nil
}

// Open a Slack modal using the views.open API
func OpenSlackModal(triggerID string, modal map[string]any, botToken string) error {
	// Construct the payload for the views.open API
	payload := map[string]any{
		"trigger_id": triggerID,
		"view":       modal,
	}

	// Make the HTTP request to Slack's API
	slackURL := "https://slack.com/api/views.open"
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Add the Authorization header with the bot token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response from Slack
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var respData map[string]any
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	// Check if the API call was successful
	if ok, okType := respData["ok"].(bool); !okType || !ok {
		metaMsg := ""
		if meta, ok := respData["response_metadata"].(map[string]any); ok {
			if msgs, ok := meta["messages"].([]any); ok && len(msgs) > 0 {
				metaMsg = fmt.Sprintf(" Details: %v", msgs)
			}
		}
		return fmt.Errorf("slack API error: %v.%s", respData["error"], metaMsg)
	}

	return nil
}

func GetAppHomeChannelID(botToken, userID string) (string, error) {
	payload := map[string]any{
		"users": []string{userID},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://slack.com/api/conversations.open", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var respBody struct {
		OK      bool `json:"ok"`
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
		Error            string `json:"error"`
		ResponseMetadata struct {
			Messages []string `json:"messages"`
		} `json:"response_metadata,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}
	if !respBody.OK {
		metaMsg := ""
		if len(respBody.ResponseMetadata.Messages) > 0 {
			metaMsg = fmt.Sprintf(" Details: %v", respBody.ResponseMetadata.Messages)
		}
		return "", fmt.Errorf("conversations.open error: %s.%s", respBody.Error, metaMsg)
	}
	return respBody.Channel.ID, nil
}
