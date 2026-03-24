package config

import (
	"os"
	"testing"
)

func TestLoad_MissingRequired(t *testing.T) {
	os.Clearenv()
	_, err := Load()
	if err == nil {
		t.Error("expected error for missing required env vars")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	os.Clearenv()
	os.Setenv("GITEA_URL", "https://git.example.com")
	os.Setenv("GITEA_TOKEN", "token123")
	os.Setenv("WEBHOOK_SECRET", "secret")
	os.Setenv("SSH_KEY_PATH", "/secrets/id_ed25519")
	os.Setenv("CLAUDE_SKILL", "pr-reviewer")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-xxx")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GiteaURL != "https://git.example.com" {
		t.Errorf("GiteaURL = %q, want %q", cfg.GiteaURL, "https://git.example.com")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080 (default)", cfg.Port)
	}
}

func TestLoad_RequiresClaudeAuth(t *testing.T) {
	os.Clearenv()
	os.Setenv("GITEA_URL", "https://git.example.com")
	os.Setenv("GITEA_TOKEN", "token123")
	os.Setenv("WEBHOOK_SECRET", "secret")
	os.Setenv("SSH_KEY_PATH", "/secrets/id_ed25519")
	os.Setenv("CLAUDE_SKILL", "pr-reviewer")
	// No ANTHROPIC_API_KEY or CLAUDE_CONFIG_PATH

	_, err := Load()
	if err == nil {
		t.Error("expected error when no Claude auth configured")
	}
}

func TestLoad_ClaudeConfigPathOnlyAuth(t *testing.T) {
	os.Clearenv()
	os.Setenv("GITEA_URL", "https://git.example.com")
	os.Setenv("GITEA_TOKEN", "token123")
	os.Setenv("WEBHOOK_SECRET", "secret")
	os.Setenv("SSH_KEY_PATH", "/secrets/id_ed25519")
	os.Setenv("CLAUDE_SKILL", "pr-reviewer")
	os.Setenv("CLAUDE_CONFIG_PATH", "/home/user/.claude.json")
	// No ANTHROPIC_API_KEY

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error with CLAUDE_CONFIG_PATH set, got: %v", err)
	}
	if cfg.ClaudeConfigPath != "/home/user/.claude.json" {
		t.Errorf("ClaudeConfigPath = %q, want %q", cfg.ClaudeConfigPath, "/home/user/.claude.json")
	}
}

func TestLoad_InvalidIntEnvVar(t *testing.T) {
	os.Clearenv()
	os.Setenv("GITEA_URL", "https://git.example.com")
	os.Setenv("GITEA_TOKEN", "token123")
	os.Setenv("WEBHOOK_SECRET", "secret")
	os.Setenv("SSH_KEY_PATH", "/secrets/id_ed25519")
	os.Setenv("CLAUDE_SKILL", "pr-reviewer")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-xxx")
	os.Setenv("RETENTION_HOURS", "24h") // invalid: should be a plain integer

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid RETENTION_HOURS value")
	}
}
