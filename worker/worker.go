package worker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kaputi/gitea-hooks/git"
	"github.com/kaputi/gitea-hooks/gitea"
	"github.com/kaputi/gitea-hooks/queue"
	"github.com/kaputi/gitea-hooks/reviewer"
)

type Worker struct {
	queue          *queue.Queue
	git            *git.Git
	reviewer       *reviewer.Reviewer
	giteaClient    *gitea.Client
	basePath       string
	retentionHours int
	logger         *slog.Logger
	wg             sync.WaitGroup
	stopCh         chan struct{}
}

func New(
	q *queue.Queue,
	g *git.Git,
	r *reviewer.Reviewer,
	gc *gitea.Client,
	basePath string,
	retentionHours int,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		queue:          q,
		git:            g,
		reviewer:       r,
		giteaClient:    gc,
		basePath:       basePath,
		retentionHours: retentionHours,
		logger:         logger,
		stopCh:         make(chan struct{}),
	}
}

func (w *Worker) Start() {
	w.wg.Add(2)
	go w.processJobs()
	go w.cleanupLoop()
}

func (w *Worker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

func (w *Worker) processJobs() {
	defer w.wg.Done()
	for {
		select {
		case job, ok := <-w.queue.Jobs():
			if !ok {
				return
			}
			w.processJob(job)
		case <-w.stopCh:
			return
		}
	}
}

func (w *Worker) processJob(job queue.Job) {
	logger := w.logger.With("repo", job.RepoName, "pr", job.PRNumber)
	logger.Info("processing job")

	workDir := w.generateWorkDir(job.RepoName, job.PRNumber)

	if err := os.MkdirAll(workDir, 0755); err != nil {
		logger.Error("failed to create work dir", "error", err)
		w.postError(job, "Review failed: internal error")
		return
	}

	success := false
	defer func() {
		if !success {
			os.RemoveAll(workDir)
		}
	}()

	clonePath := filepath.Join(workDir, job.RepoName)
	if err := w.git.Clone(job.RepoURL, job.Branch, clonePath); err != nil {
		logger.Error("clone failed", "error", err)
		w.postError(job, "Review failed: could not clone repo")
		return
	}

	review, err := w.reviewer.Review(clonePath)
	if err != nil {
		logger.Error("review failed", "error", err)
		w.postError(job, "Review failed: reviewer error")
		return
	}

	if err := w.giteaClient.PostComment(job.RepoOwner, job.RepoName, job.PRNumber, review); err != nil {
		logger.Error("failed to post comment", "error", err)
		w.postError(job, "Review completed but failed to post comment")
		return
	}

	success = true
	logger.Info("job completed successfully")
}

func (w *Worker) postError(job queue.Job, message string) {
	if err := w.giteaClient.PostComment(job.RepoOwner, job.RepoName, job.PRNumber, message); err != nil {
		w.logger.Error("failed to post error comment", "error", err)
	}
}

func (w *Worker) generateWorkDir(repoName string, prNumber int) string {
	timestamp := time.Now().Unix()
	return filepath.Join(w.basePath, fmt.Sprintf("%s-%d-%d", repoName, prNumber, timestamp))
}

func (w *Worker) cleanupLoop() {
	defer w.wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.cleanup()
		case <-w.stopCh:
			return
		}
	}
}

func (w *Worker) cleanup() {
	entries, err := os.ReadDir(w.basePath)
	if err != nil {
		w.logger.Warn("cleanup: failed to read base path", "error", err)
		return
	}

	cutoff := time.Now().Add(-time.Duration(w.retentionHours) * time.Hour)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(w.basePath, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				w.logger.Warn("cleanup: failed to remove dir", "path", path, "error", err)
			} else {
				w.logger.Info("cleanup: removed old directory", "path", path)
			}
		}
	}
}
