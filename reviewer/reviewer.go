package reviewer

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type Reviewer struct {
	skill  string
	apiKey string
}

func New(skill, apiKey string) *Reviewer {
	return &Reviewer{
		skill:  skill,
		apiKey: apiKey,
	}
}

func (r *Reviewer) Review(repoPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	args := r.buildArgs()
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = repoPath

	if r.apiKey != "" {
		cmd.Env = append(cmd.Environ(), "ANTHROPIC_API_KEY="+r.apiKey)
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude failed: %w\nstderr: %s", err, exitErr.Stderr)
		}
		return "", fmt.Errorf("claude failed: %w", err)
	}
	return string(output), nil
}

func (r *Reviewer) buildArgs() []string {
	return []string{
		"--print",
		"/" + r.skill,
	}
}
