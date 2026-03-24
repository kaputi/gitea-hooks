package git

import (
	"strings"
	"testing"
)

func TestBuildCloneCommand(t *testing.T) {
	g := New("/path/to/key")
	args := g.buildCloneArgs("git@example.com:user/repo.git", "feature", "/tmp/clone")

	expected := []string{
		"clone",
		"--branch", "feature",
		"--single-branch",
		"--depth", "1",
		"git@example.com:user/repo.git",
		"/tmp/clone",
	}

	if len(args) != len(expected) {
		t.Fatalf("args length = %d, want %d", len(args), len(expected))
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestSSHCommand(t *testing.T) {
	g := New("/secrets/id_ed25519")
	sshCmd := g.sshCommand()

	if !strings.Contains(sshCmd, "/secrets/id_ed25519") {
		t.Errorf("SSH command should reference key path: %s", sshCmd)
	}
	if !strings.Contains(sshCmd, "StrictHostKeyChecking=accept-new") {
		t.Errorf("SSH command should accept new host keys: %s", sshCmd)
	}
}

func TestSSHCommandPathWithSpaces(t *testing.T) {
	g := New("/home/user/my keys/id_ed25519")
	sshCmd := g.sshCommand()

	expected := "ssh -i '/home/user/my keys/id_ed25519' -o StrictHostKeyChecking=accept-new"
	if sshCmd != expected {
		t.Errorf("SSH command = %q, want %q", sshCmd, expected)
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/simple/path", "'/simple/path'"},
		{"/path/with spaces/key", "'/path/with spaces/key'"},
		{"it's a trap", "'it'\"'\"'s a trap'"},
	}

	for _, tt := range tests {
		got := shellQuote(tt.input)
		if got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
