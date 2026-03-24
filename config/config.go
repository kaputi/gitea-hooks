package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// Required
	GiteaURL      string
	GiteaToken    string
	WebhookSecret string
	SSHKeyPath    string
	ClaudeSkill   string

	// Claude auth (one required)
	AnthropicAPIKey  string
	ClaudeConfigPath string

	// Optional with defaults
	Port           int
	CloneBasePath  string
	RetentionHours int
	QueueSize      int
}

func Load() (*Config, error) {
	port, err := getEnvInt("PORT", 8080)
	if err != nil {
		return nil, err
	}
	retentionHours, err := getEnvInt("RETENTION_HOURS", 24)
	if err != nil {
		return nil, err
	}
	queueSize, err := getEnvInt("QUEUE_SIZE", 100)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		GiteaURL:         os.Getenv("GITEA_URL"),
		GiteaToken:       os.Getenv("GITEA_TOKEN"),
		WebhookSecret:    os.Getenv("WEBHOOK_SECRET"),
		SSHKeyPath:       os.Getenv("SSH_KEY_PATH"),
		ClaudeSkill:      os.Getenv("CLAUDE_SKILL"),
		AnthropicAPIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		ClaudeConfigPath: os.Getenv("CLAUDE_CONFIG_PATH"),
		Port:             port,
		CloneBasePath:    getEnvString("CLONE_BASE_PATH", "/data/reviews"),
		RetentionHours:   retentionHours,
		QueueSize:        queueSize,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.GiteaURL == "" {
		return errors.New("GITEA_URL is required")
	}
	if c.GiteaToken == "" {
		return errors.New("GITEA_TOKEN is required")
	}
	if c.WebhookSecret == "" {
		return errors.New("WEBHOOK_SECRET is required")
	}
	if c.SSHKeyPath == "" {
		return errors.New("SSH_KEY_PATH is required")
	}
	if c.ClaudeSkill == "" {
		return errors.New("CLAUDE_SKILL is required")
	}
	if c.AnthropicAPIKey == "" && c.ClaudeConfigPath == "" {
		return errors.New("either ANTHROPIC_API_KEY or CLAUDE_CONFIG_PATH is required")
	}
	return nil
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer, got %q", key, val)
	}
	return i, nil
}
