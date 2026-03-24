package queue

import (
	"errors"
	"sync"
)

var (
	ErrQueueFull   = errors.New("queue is full")
	ErrQueueClosed = errors.New("queue is closed")
)

type Job struct {
	RepoURL   string
	Branch    string
	PRNumber  int
	RepoOwner string
	RepoName  string
}

type Queue struct {
	jobs   chan Job
	closed bool
	mu     sync.Mutex
	once   sync.Once
}

func New(size int) *Queue {
	return &Queue{
		jobs: make(chan Job, size),
	}
}

func (q *Queue) Enqueue(job Job) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return ErrQueueClosed
	}
	q.mu.Unlock()

	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}

func (q *Queue) Close() {
	q.once.Do(func() {
		q.mu.Lock()
		q.closed = true
		q.mu.Unlock()
		close(q.jobs)
	})
}
