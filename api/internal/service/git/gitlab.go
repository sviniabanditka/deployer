package git

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// VerifyGitLabToken verifies a GitLab webhook token.
func VerifyGitLabToken(secret, token string) bool {
	return subtle.ConstantTimeCompare([]byte(secret), []byte(token)) == 1
}

type gitLabPushEvent struct {
	Ref    string `json:"ref"`
	After  string `json:"after"`
}

// ParseGitLabPushEvent extracts branch and commit SHA from a GitLab push event payload.
func ParseGitLabPushEvent(payload []byte) (branch string, commitSHA string, err error) {
	var event gitLabPushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return "", "", fmt.Errorf("failed to parse gitlab push event: %w", err)
	}

	if !strings.HasPrefix(event.Ref, "refs/heads/") {
		return "", "", fmt.Errorf("not a branch push: %s", event.Ref)
	}

	branch = strings.TrimPrefix(event.Ref, "refs/heads/")
	return branch, event.After, nil
}

type createGitLabWebhookRequest struct {
	URL                   string `json:"url"`
	Token                 string `json:"token"`
	PushEvents            bool   `json:"push_events"`
	MergeRequestsEvents   bool   `json:"merge_requests_events"`
	EnableSSLVerification bool   `json:"enable_ssl_verification"`
}

type createGitLabWebhookResponse struct {
	ID int64 `json:"id"`
}

// CreateGitLabWebhook registers a webhook on a GitLab project.
// projectID should be the URL-encoded path or numeric ID of the project.
func CreateGitLabWebhook(accessToken, projectID, webhookURL, secret string) (int64, error) {
	reqBody := createGitLabWebhookRequest{
		URL:                   webhookURL,
		Token:                 secret,
		PushEvents:            true,
		MergeRequestsEvents:   true,
		EnableSSLVerification: true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal webhook request: %w", err)
	}

	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/hooks", url.PathEscape(projectID))
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to create gitlab webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("gitlab API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var webhookResp createGitLabWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhookResp); err != nil {
		return 0, fmt.Errorf("failed to decode gitlab webhook response: %w", err)
	}

	return webhookResp.ID, nil
}

type gitLabMREvent struct {
	ObjectKind       string `json:"object_kind"` // merge_request
	ObjectAttributes struct {
		Action      string `json:"action"` // open, update, close, merge
		IID         int    `json:"iid"`
		URL         string `json:"url"`
		SourceBranch string `json:"source_branch"`
	} `json:"object_attributes"`
	Project struct {
		ID int `json:"id"`
	} `json:"project"`
}

// ParseGitLabMREvent extracts action, MR number, MR URL, and source branch from a GitLab merge_request event payload.
func ParseGitLabMREvent(payload []byte) (action string, mrNumber int, mrURL string, sourceBranch string, projectID int, err error) {
	var event gitLabMREvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return "", 0, "", "", 0, fmt.Errorf("failed to parse gitlab MR event: %w", err)
	}

	return event.ObjectAttributes.Action, event.ObjectAttributes.IID,
		event.ObjectAttributes.URL, event.ObjectAttributes.SourceBranch,
		event.Project.ID, nil
}

// PostGitLabComment posts a comment on a GitLab merge request.
func PostGitLabComment(accessToken string, projectID int, mrNumber int, body string) error {
	commentBody := map[string]string{"body": body}
	bodyBytes, err := json.Marshal(commentBody)
	if err != nil {
		return fmt.Errorf("failed to marshal comment body: %w", err)
	}

	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%d/merge_requests/%d/notes", projectID, mrNumber)
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post gitlab comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteGitLabWebhook removes a webhook from a GitLab project.
func DeleteGitLabWebhook(accessToken, projectID string, webhookID int64) error {
	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/hooks/%d", url.PathEscape(projectID), webhookID)
	req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete gitlab webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
