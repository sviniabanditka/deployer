package git

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// VerifyGitHubSignature verifies a GitHub webhook signature (SHA-256).
func VerifyGitHubSignature(secret, signature string, body []byte) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	sig := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expected))
}

type gitHubPushEvent struct {
	Ref        string `json:"ref"`
	HeadCommit struct {
		ID string `json:"id"`
	} `json:"head_commit"`
}

// ParseGitHubPushEvent extracts branch and commit SHA from a GitHub push event payload.
func ParseGitHubPushEvent(payload []byte) (branch string, commitSHA string, err error) {
	var event gitHubPushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return "", "", fmt.Errorf("failed to parse github push event: %w", err)
	}

	if !strings.HasPrefix(event.Ref, "refs/heads/") {
		return "", "", fmt.Errorf("not a branch push: %s", event.Ref)
	}

	branch = strings.TrimPrefix(event.Ref, "refs/heads/")
	return branch, event.HeadCommit.ID, nil
}

type createGitHubWebhookRequest struct {
	Name   string                 `json:"name"`
	Active bool                   `json:"active"`
	Events []string               `json:"events"`
	Config map[string]interface{} `json:"config"`
}

type createGitHubWebhookResponse struct {
	ID int64 `json:"id"`
}

// CreateGitHubWebhook registers a webhook on a GitHub repository.
func CreateGitHubWebhook(accessToken, owner, repo, webhookURL, secret string) (int64, error) {
	reqBody := createGitHubWebhookRequest{
		Name:   "web",
		Active: true,
		Events: []string{"push", "pull_request"},
		Config: map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       secret,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal webhook request: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to create github webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var webhookResp createGitHubWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhookResp); err != nil {
		return 0, fmt.Errorf("failed to decode github webhook response: %w", err)
	}

	return webhookResp.ID, nil
}

type gitHubPREvent struct {
	Action      string `json:"action"` // opened, synchronize, closed
	Number      int    `json:"number"`
	PullRequest struct {
		HTMLURL string `json:"html_url"`
		Head    struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
}

// ParseGitHubPREvent extracts action, PR number, PR URL, head branch, and head SHA from a GitHub pull_request event payload.
func ParseGitHubPREvent(payload []byte) (action string, prNumber int, prURL string, headBranch string, headSHA string, err error) {
	var event gitHubPREvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return "", 0, "", "", "", fmt.Errorf("failed to parse github PR event: %w", err)
	}

	return event.Action, event.Number, event.PullRequest.HTMLURL,
		event.PullRequest.Head.Ref, event.PullRequest.Head.SHA, nil
}

// PostGitHubComment posts a comment on a GitHub pull request.
func PostGitHubComment(accessToken, owner, repo string, prNumber int, body string) error {
	commentBody := map[string]string{"body": body}
	bodyBytes, err := json.Marshal(commentBody)
	if err != nil {
		return fmt.Errorf("failed to marshal comment body: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post github comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteGitHubWebhook removes a webhook from a GitHub repository.
func DeleteGitHubWebhook(accessToken, owner, repo string, webhookID int64) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks/%d", owner, repo, webhookID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete github webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
