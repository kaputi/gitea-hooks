package reviewer

import (
	"testing"
)

func TestBuildCommand(t *testing.T) {
	r := New("pr-reviewer", "")
	args := r.buildArgs()

	if args[0] != "--print" {
		t.Errorf("args[0] = %q, want '--print'", args[0])
	}
	if args[1] != "/pr-reviewer" {
		t.Errorf("args[1] = %q, want '/pr-reviewer'", args[1])
	}
}

func TestBuildCommandWithPrompt(t *testing.T) {
	r := New("pr-reviewer", "")
	args := r.buildArgs()

	found := false
	for _, arg := range args {
		if arg == "/pr-reviewer" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("skill name not found in args: %v", args)
	}
}
