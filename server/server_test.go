package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kaputi/gitea-hooks/queue"
)

func TestHealthEndpoint(t *testing.T) {
	q := queue.New(10)
	defer q.Close()
	s := New(q, "secret", nil)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestWebhookEndpoint_InvalidSignature(t *testing.T) {
	q := queue.New(10)
	defer q.Close()
	s := New(q, "secret", nil)

	body := `{"action":"opened"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Gitea-Signature", "invalid")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestWebhookEndpoint_ValidRequest(t *testing.T) {
	q := queue.New(10)
	defer q.Close()
	s := New(q, "secret", nil)

	body := `{
		"action": "opened",
		"number": 1,
		"pull_request": {"head": {"ref": "branch"}},
		"repository": {
			"owner": {"login": "owner"},
			"name": "repo",
			"ssh_url": "git@example.com:owner/repo.git"
		}
	}`

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(body))
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Gitea-Signature", signature)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
