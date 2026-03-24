package queue

import (
	"errors"
	"testing"
	"time"
)

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := New(10)
	defer q.Close()

	job := Job{
		RepoURL:   "git@example.com:user/repo.git",
		Branch:    "feature-branch",
		PRNumber:  42,
		RepoOwner: "user",
		RepoName:  "repo",
	}

	if err := q.Enqueue(job); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	select {
	case received := <-q.Jobs():
		if received.PRNumber != 42 {
			t.Errorf("PRNumber = %d, want 42", received.PRNumber)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for job")
	}
}

func TestQueue_DoubleClose(t *testing.T) {
	q := New(10)
	// Neither call should panic.
	q.Close()
	q.Close()
}

func TestQueue_EnqueueAfterClose(t *testing.T) {
	q := New(10)
	q.Close()

	err := q.Enqueue(Job{PRNumber: 1})
	if !errors.Is(err, ErrQueueClosed) {
		t.Errorf("expected ErrQueueClosed, got %v", err)
	}
}

func TestQueue_Full(t *testing.T) {
	q := New(1)
	defer q.Close()

	job := Job{PRNumber: 1}
	if err := q.Enqueue(job); err != nil {
		t.Fatalf("first enqueue failed: %v", err)
	}

	// Queue is full, should return ErrQueueFull
	err := q.Enqueue(Job{PRNumber: 2})
	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}
