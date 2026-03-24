package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrIgnoredAction    = errors.New("action ignored")
	ErrEmptySecret      = errors.New("webhook secret must not be empty")
)

type PREvent struct {
	Action    string
	PRNumber  int
	Branch    string
	RepoURL   string
	RepoOwner string
	RepoName  string
}

type giteaPayload struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Head struct {
			Ref string `json:"ref"`
		} `json:"head"`
	} `json:"pull_request"`
	Repository struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name   string `json:"name"`
		SSHURL string `json:"ssh_url"`
	} `json:"repository"`
}

func ValidateSignature(body []byte, signature, secret string) error {
	if secret == "" {
		return ErrEmptySecret
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return ErrInvalidSignature
	}
	return nil
}

func ParsePREvent(body []byte) (*PREvent, error) {
	var payload giteaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	if payload.Action != "opened" {
		return nil, ErrIgnoredAction
	}

	return &PREvent{
		Action:    payload.Action,
		PRNumber:  payload.Number,
		Branch:    payload.PullRequest.Head.Ref,
		RepoURL:   payload.Repository.SSHURL,
		RepoOwner: payload.Repository.Owner.Login,
		RepoName:  payload.Repository.Name,
	}, nil
}
