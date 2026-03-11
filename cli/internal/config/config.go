package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIUrl       string `json:"api_url,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Email        string `json:"email,omitempty"`
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".deployer", "config.json")
	}
	return filepath.Join(home, ".deployer", "config.json")
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadAppConfig reads .deployer.json from the given directory and returns the app ID.
func LoadAppConfig(dir string) (string, error) {
	path := filepath.Join(dir, ".deployer.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var appCfg struct {
		AppID string `json:"app_id"`
	}
	if err := json.Unmarshal(data, &appCfg); err != nil {
		return "", err
	}
	return appCfg.AppID, nil
}

// SaveAppConfig writes .deployer.json to the given directory.
func SaveAppConfig(dir string, appID string) error {
	path := filepath.Join(dir, ".deployer.json")
	data, err := json.MarshalIndent(map[string]string{"app_id": appID}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
