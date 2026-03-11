package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/deployer/cli/internal/config"
)

type Client struct {
	BaseURL      string
	AccessToken  string
	RefreshToken string
	ConfigPath   string
	HTTPClient   *http.Client
}

func New(baseURL, accessToken, refreshToken, configPath string) *Client {
	return &Client{
		BaseURL:      strings.TrimRight(baseURL, "/"),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ConfigPath:   configPath,
		HTTPClient:   &http.Client{},
	}
}

// ---------- Auth ----------

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
}

func (c *Client) Login(email, password string) (*AuthResponse, error) {
	body := LoginRequest{Email: email, Password: password}
	var resp AuthResponse
	if err := c.doJSON("POST", "/auth/login", body, &resp); err != nil {
		return nil, err
	}
	c.AccessToken = resp.AccessToken
	c.RefreshToken = resp.RefreshToken
	return &resp, nil
}

func (c *Client) Register(name, email, password string) (*AuthResponse, error) {
	body := RegisterRequest{Name: name, Email: email, Password: password}
	var resp AuthResponse
	if err := c.doJSON("POST", "/auth/register", body, &resp); err != nil {
		return nil, err
	}
	c.AccessToken = resp.AccessToken
	c.RefreshToken = resp.RefreshToken
	return &resp, nil
}

// ---------- Apps ----------

type App struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Slug      string            `json:"slug"`
	Status    string            `json:"status"`
	EnvVars   map[string]string `json:"env_vars,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

type CreateAppRequest struct {
	Name string `json:"name"`
}

func (c *Client) CreateApp(name string) (*App, error) {
	var app App
	if err := c.doJSON("POST", "/apps", CreateAppRequest{Name: name}, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *Client) ListApps() ([]App, error) {
	var apps []App
	if err := c.doJSON("GET", "/apps", nil, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *Client) GetApp(id string) (*App, error) {
	var app App
	if err := c.doJSON("GET", "/apps/"+id, nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *Client) DeleteApp(id string) error {
	return c.doJSON("DELETE", "/apps/"+id, nil, nil)
}

func (c *Client) StopApp(id string) error {
	return c.doJSON("POST", "/apps/"+id+"/stop", nil, nil)
}

func (c *Client) StartApp(id string) error {
	return c.doJSON("POST", "/apps/"+id+"/start", nil, nil)
}

func (c *Client) UpdateEnvVars(id string, envVars map[string]string) error {
	body := map[string]interface{}{"env_vars": envVars}
	return c.doJSON("PUT", "/apps/"+id+"/env", body, nil)
}

// ---------- Databases ----------

type Database struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Engine        string `json:"engine"`
	Version       string `json:"version,omitempty"`
	Status        string `json:"status"`
	AppID         string `json:"app_id,omitempty"`
	ConnectionURL string `json:"connection_url,omitempty"`
	Host          string `json:"host,omitempty"`
	Port          int    `json:"port,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	DatabaseName  string `json:"database_name,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
}

type Backup struct {
	ID        string `json:"id"`
	DatabaseID string `json:"database_id,omitempty"`
	Status    string `json:"status"`
	Size      string `json:"size,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type CreateDatabaseRequest struct {
	Name   string `json:"name"`
	Engine string `json:"engine"`
	Version string `json:"version,omitempty"`
	AppID  string `json:"app_id,omitempty"`
}

func (c *Client) CreateDatabase(name, engine, version string, appID string) (*Database, error) {
	body := CreateDatabaseRequest{Name: name, Engine: engine, Version: version, AppID: appID}
	var db Database
	if err := c.doJSON("POST", "/databases", body, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (c *Client) ListDatabases() ([]Database, error) {
	var dbs []Database
	if err := c.doJSON("GET", "/databases", nil, &dbs); err != nil {
		return nil, err
	}
	return dbs, nil
}

func (c *Client) GetDatabase(id string) (*Database, error) {
	var db Database
	if err := c.doJSON("GET", "/databases/"+id, nil, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (c *Client) DeleteDatabase(id string) error {
	return c.doJSON("DELETE", "/databases/"+id, nil, nil)
}

func (c *Client) StopDatabase(id string) error {
	return c.doJSON("POST", "/databases/"+id+"/stop", nil, nil)
}

func (c *Client) StartDatabase(id string) error {
	return c.doJSON("POST", "/databases/"+id+"/start", nil, nil)
}

func (c *Client) LinkDatabase(dbID, appID string) error {
	body := map[string]string{"app_id": appID}
	return c.doJSON("POST", "/databases/"+dbID+"/link", body, nil)
}

func (c *Client) UnlinkDatabase(dbID string) error {
	return c.doJSON("POST", "/databases/"+dbID+"/unlink", nil, nil)
}

func (c *Client) CreateBackup(dbID string) (*Backup, error) {
	var backup Backup
	if err := c.doJSON("POST", "/databases/"+dbID+"/backups", nil, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

func (c *Client) ListBackups(dbID string) ([]Backup, error) {
	var backups []Backup
	if err := c.doJSON("GET", "/databases/"+dbID+"/backups", nil, &backups); err != nil {
		return nil, err
	}
	return backups, nil
}

func (c *Client) RestoreBackup(dbID, backupID string) error {
	return c.doJSON("POST", "/databases/"+dbID+"/backups/"+backupID+"/restore", nil, nil)
}

// ---------- Deploy ----------

type DeployResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (c *Client) Deploy(appID, zipPath string) (*DeployResponse, error) {
	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to write file to form: %w", err)
	}
	writer.Close()

	url := c.BaseURL + "/apps/" + appID + "/deploy"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		if refreshErr := c.tryRefreshToken(); refreshErr == nil {
			// Retry once after token refresh
			file.Seek(0, 0)
			var buf2 bytes.Buffer
			writer2 := multipart.NewWriter(&buf2)
			part2, _ := writer2.CreateFormFile("file", filepath.Base(zipPath))
			io.Copy(part2, file)
			writer2.Close()

			req2, _ := http.NewRequest("POST", url, &buf2)
			req2.Header.Set("Content-Type", writer2.FormDataContentType())
			req2.Header.Set("Authorization", "Bearer "+c.AccessToken)
			resp2, err := c.HTTPClient.Do(req2)
			if err != nil {
				return nil, fmt.Errorf("request failed after token refresh: %w", err)
			}
			defer resp2.Body.Close()
			resp = resp2
		}
	}

	if resp.StatusCode >= 400 {
		return nil, parseError(resp)
	}

	var deployResp DeployResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &deployResp, nil
}

// ---------- Logs ----------

func (c *Client) GetLogs(appID string) (string, error) {
	respBody, err := c.doRaw("GET", "/apps/"+appID+"/logs")
	if err != nil {
		return "", err
	}
	defer respBody.Close()
	data, err := io.ReadAll(respBody)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) LogsWebSocketURL(appID string) string {
	wsURL := strings.Replace(c.BaseURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	return wsURL + "/apps/" + appID + "/logs?stream=true&token=" + c.AccessToken
}

// ---------- Deployment Status ----------

type DeploymentStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Logs   string `json:"logs,omitempty"`
}

func (c *Client) GetDeploymentStatus(appID, deploymentID string) (*DeploymentStatus, error) {
	var status DeploymentStatus
	if err := c.doJSON("GET", "/apps/"+appID+"/deployments/"+deploymentID, nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// ---------- Internal ----------

func (c *Client) doJSON(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		if refreshErr := c.tryRefreshToken(); refreshErr == nil {
			// Retry the request with the new token
			var retryReader io.Reader
			if body != nil {
				data, _ := json.Marshal(body)
				retryReader = bytes.NewReader(data)
			}
			req2, _ := http.NewRequest(method, url, retryReader)
			if body != nil {
				req2.Header.Set("Content-Type", "application/json")
			}
			req2.Header.Set("Authorization", "Bearer "+c.AccessToken)
			resp2, err := c.HTTPClient.Do(req2)
			if err != nil {
				return fmt.Errorf("request failed after token refresh: %w", err)
			}
			defer resp2.Body.Close()
			resp = resp2
		}
	}

	if resp.StatusCode >= 400 {
		return parseError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) doRaw(method, path string) (io.ReadCloser, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == 401 {
		resp.Body.Close()
		if refreshErr := c.tryRefreshToken(); refreshErr == nil {
			req2, _ := http.NewRequest(method, url, nil)
			req2.Header.Set("Authorization", "Bearer "+c.AccessToken)
			resp2, err := c.HTTPClient.Do(req2)
			if err != nil {
				return nil, fmt.Errorf("request failed after token refresh: %w", err)
			}
			resp = resp2
		}
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseError(resp)
	}

	return resp.Body, nil
}

func (c *Client) tryRefreshToken() error {
	if c.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	body := map[string]string{"refresh_token": c.RefreshToken}
	data, _ := json.Marshal(body)
	url := c.BaseURL + "/auth/refresh"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("token refresh failed")
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	c.AccessToken = authResp.AccessToken
	c.RefreshToken = authResp.RefreshToken

	// Persist refreshed tokens
	cfg, err := config.Load(c.ConfigPath)
	if err == nil {
		cfg.AccessToken = c.AccessToken
		cfg.RefreshToken = c.RefreshToken
		config.Save(c.ConfigPath, cfg)
	}

	return nil
}

type APIError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func parseError(resp *http.Response) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var apiErr APIError
	if err := json.Unmarshal(data, &apiErr); err == nil {
		msg := apiErr.Message
		if msg == "" {
			msg = apiErr.Error
		}
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}

	return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(data))
}
