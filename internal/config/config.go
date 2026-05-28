package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey  string
	BaseURL string
	Model   string

	TTS TTSConfig
}

type TTSConfig struct {
	APIKey     string
	ResourceID string
	VoiceType  string
	Endpoint   string
}

type settingsFile struct {
	Env map[string]string `json:"env"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".mooncli", "settings.json"))
	if err != nil {
		return nil, fmt.Errorf("read settings.json: %w", err)
	}

	var sf settingsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parse settings.json: %w", err)
	}

	cfg := &Config{
		APIKey:  sf.Env["ANTHROPIC_AUTH_TOKEN"],
		BaseURL: sf.Env["ANTHROPIC_BASE_URL"],
		Model:   sf.Env["ANTHROPIC_MODEL"],
		TTS: TTSConfig{
			APIKey:     sf.Env["VOLCENGINE_TTS_API_KEY"],
			ResourceID: sf.Env["VOLCENGINE_TTS_RESOURCE_ID"],
			VoiceType:  sf.Env["VOLCENGINE_TTS_VOICE_TYPE"],
			Endpoint:   sf.Env["VOLCENGINE_TTS_ENDPOINT"],
		},
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_AUTH_TOKEN not set in settings.json")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("ANTHROPIC_BASE_URL not set in settings.json")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("ANTHROPIC_MODEL not set in settings.json")
	}

	return cfg, nil
}
