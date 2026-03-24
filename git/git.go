package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Git struct {
	sshKeyPath string
}

func New(sshKeyPath string) *Git {
	return &Git{sshKeyPath: sshKeyPath}
}

func (g *Git) Clone(repoURL, branch, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := g.buildCloneArgs(repoURL, branch, destPath)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(cmd.Environ(), "GIT_SSH_COMMAND="+g.sshCommand())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\noutput: %s", err, output)
	}
	return nil
}

func (g *Git) buildCloneArgs(repoURL, branch, destPath string) []string {
	return []string{
		"clone",
		"--branch", branch,
		"--single-branch",
		"--depth", "1",
		repoURL,
		destPath,
	}
}

func (g *Git) sshCommand() string {
	return fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=accept-new", shellQuote(g.sshKeyPath))
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
