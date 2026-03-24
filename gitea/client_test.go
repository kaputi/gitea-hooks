package gitea

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_PostComment(t *testing.T) {
	var receivedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/repos/owner/repo/issues/42/comments" {
			t.Errorf("path = %s, want /api/v1/repos/owner/repo/issues/42/comments", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("auth header = %s, want 'token test-token'", r.Header.Get("Authorization"))
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(server.URL, "test-token")
	err := client.PostComment("owner", "repo", 42, "Review comment")
	if err != nil {
		t.Fatalf("PostComment failed: %v", err)
	}
	if receivedBody["body"] != "Review comment" {
		t.Errorf("body = %q, want %q", receivedBody["body"], "Review comment")
	}
}

func TestClient_PostComment_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := New(server.URL, "bad-token")
	err := client.PostComment("owner", "repo", 1, "comment")
	if err == nil {
		t.Fatal("PostComment should have returned an error on 401")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (4xx must not be retried)", attempts)
	}
}

func TestClient_PostComment_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(server.URL, "test-token")
	err := client.PostComment("owner", "repo", 1, "comment")
	if err != nil {
		t.Fatalf("PostComment failed after retries: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}
