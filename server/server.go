package server

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/kaputi/gitea-hooks/queue"
	"github.com/kaputi/gitea-hooks/webhook"
)

type Server struct {
	mux    *http.ServeMux
	queue  *queue.Queue
	secret string
	logger *slog.Logger
}

func New(q *queue.Queue, secret string, logger *slog.Logger) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		queue:  q,
		secret: secret,
		logger: logger,
	}
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /webhook", s.handleWebhook)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Gitea-Signature")
	if err := webhook.ValidateSignature(body, signature, s.secret); err != nil {
		if s.logger != nil {
			s.logger.Warn("invalid webhook signature")
		}
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	event, err := webhook.ParsePREvent(body)
	if err == webhook.ErrIgnoredAction {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to parse webhook", "error", err)
		}
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	job := queue.Job{
		RepoURL:   event.RepoURL,
		Branch:    event.Branch,
		PRNumber:  event.PRNumber,
		RepoOwner: event.RepoOwner,
		RepoName:  event.RepoName,
	}

	if err := s.queue.Enqueue(job); err != nil {
		if s.logger != nil {
			s.logger.Warn("queue full", "error", err)
		}
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	if s.logger != nil {
		s.logger.Info("job enqueued", "repo", event.RepoName, "pr", event.PRNumber)
	}
	w.WriteHeader(http.StatusOK)
}
