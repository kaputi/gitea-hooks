package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func TestValidateSignature_Valid(t *testing.T) {
	secret := "mysecret"
	body := []byte(`{"action":"opened"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	err := ValidateSignature(body, signature, secret)
	if err != nil {
		t.Errorf("ValidateSignature failed: %v", err)
	}
}

func TestValidateSignature_Invalid(t *testing.T) {
	err := ValidateSignature([]byte("body"), "invalid", "secret")
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidateSignature_EmptySecret(t *testing.T) {
	err := ValidateSignature([]byte("body"), "anysig", "")
	if !errors.Is(err, ErrEmptySecret) {
		t.Errorf("expected ErrEmptySecret, got %v", err)
	}
}

func TestParsePREvent_Opened(t *testing.T) {
	payload := []byte(`{
		"action": "opened",
		"number": 42,
		"pull_request": {
			"head": {
				"ref": "feature-branch"
			}
		},
		"repository": {
			"owner": {"login": "testowner"},
			"name": "testrepo",
			"ssh_url": "git@example.com:testowner/testrepo.git"
		}
	}`)

	event, err := ParsePREvent(payload)
	if err != nil {
		t.Fatalf("ParsePREvent failed: %v", err)
	}
	if event.Action != "opened" {
		t.Errorf("Action = %q, want 'opened'", event.Action)
	}
	if event.PRNumber != 42 {
		t.Errorf("PRNumber = %d, want 42", event.PRNumber)
	}
	if event.Branch != "feature-branch" {
		t.Errorf("Branch = %q, want 'feature-branch'", event.Branch)
	}
	if event.RepoOwner != "testowner" {
		t.Errorf("RepoOwner = %q, want 'testowner'", event.RepoOwner)
	}
}

func TestParsePREvent_IgnoreNonOpened(t *testing.T) {
	payload := []byte(`{"action": "closed", "number": 1}`)
	_, err := ParsePREvent(payload)
	if err != ErrIgnoredAction {
		t.Errorf("expected ErrIgnoredAction, got %v", err)
	}
}
