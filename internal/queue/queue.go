package queue

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/AlexandrKudryavtsev/go-job-queue/internal/job"
	"github.com/google/uuid"
)

type Config struct {
	VisibilityTimeout time.Duration
	RetryBaseDelay    time.Duration
	SweepInterval     time.Duration
	MaxPayloadSize    int
}

type Queue struct {
	mu    sync.Mutex
	jobs  map[string]*job.Job
	order []string
	cfg   Config
}

func New(cfg Config) *Queue {
	return &Queue{
		jobs:  make(map[string]*job.Job),
		order: make([]string, 0),
		cfg:   cfg,
	}
}

func cloneJob(j job.Job) job.Job {
	j.Payload = slices.Clone(j.Payload)
	return j
}

type CreateJobInput struct {
	Type         string
	Payload      json.RawMessage
	MaxAttempts  int
	DelaySeconds int
}

func (q *Queue) Create(input CreateJobInput) (job.Job, error) {
	jobType := strings.TrimSpace(input.Type)
	if jobType == "" {
		return job.Job{}, ErrInvalidJobType
	}
	if input.DelaySeconds < 0 {
		return job.Job{}, ErrInvalidDelay
	}
	if input.MaxAttempts <= 0 {
		return job.Job{}, ErrInvalidMaxAttempts
	}
	payloadLen := len(input.Payload)
	if payloadLen == 0 || payloadLen > q.cfg.MaxPayloadSize {
		return job.Job{}, ErrInvalidPayload
	}

	now := time.Now()
	id := uuid.NewString()

	payloadCopy := slices.Clone(input.Payload)

	newJob := job.Job{
		ID:          id,
		Status:      job.StatusQueued,
		Type:        jobType,
		Payload:     payloadCopy,
		MaxAttempts: input.MaxAttempts,
		CreatedAt:   now,
		AvailableAt: now.Add(time.Duration(input.DelaySeconds) * time.Second),
	}

	if input.DelaySeconds > 0 {
		newJob.Status = job.StatusDelayed
	}

	q.mu.Lock()
	q.jobs[id] = &newJob
	q.order = append(q.order, id)
	q.mu.Unlock()

	return cloneJob(newJob), nil
}

func (q *Queue) Get(id string) (job.Job, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return job.Job{}, ErrJobNotFound
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	foundJob, ok := q.jobs[id]

	if !ok {
		return job.Job{}, ErrJobNotFound
	}

	return cloneJob(*foundJob), nil
}

func (q *Queue) Next() (job.Job, error) {
	now := time.Now()

	q.mu.Lock()
	defer q.mu.Unlock()

	for _, jobID := range q.order {
		currentJob, ok := q.jobs[jobID]
		if !ok {
			continue
		}

		if currentJob.Status != job.StatusQueued && currentJob.Status != job.StatusDelayed {
			continue
		}

		if currentJob.AvailableAt.After(now) {
			continue
		}

		currentJob.Attempts++
		currentJob.StartedAt = &now
		currentJob.Status = job.StatusProcessing

		return cloneJob(*currentJob), nil
	}

	return job.Job{}, ErrNoJobsAvailable
}

func (q *Queue) Ack(jobID string) (job.Job, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return job.Job{}, ErrJobNotFound
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	currentJob, ok := q.jobs[jobID]

	if !ok {
		return job.Job{}, ErrJobNotFound
	}

	if currentJob.Status != job.StatusProcessing {
		return job.Job{}, ErrInvalidStatus
	}

	now := time.Now()

	currentJob.Status = job.StatusDone
	currentJob.FinishedAt = &now

	return cloneJob(*currentJob), nil
}

func (q *Queue) Nack(jobID, reason string) (job.Job, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return job.Job{}, ErrJobNotFound
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	currentJob, ok := q.jobs[jobID]
	if !ok {
		return job.Job{}, ErrJobNotFound
	}

	if currentJob.Status != job.StatusProcessing {
		return job.Job{}, ErrInvalidStatus
	}

	currentJob.LastError = strings.TrimSpace(reason)
	currentJob.StartedAt = nil

	if currentJob.Attempts < currentJob.MaxAttempts {
		currentJob.Status = job.StatusDelayed
		currentJob.FinishedAt = nil
		currentJob.AvailableAt = now.Add(time.Duration(currentJob.Attempts) * q.cfg.RetryBaseDelay)
	} else {
		currentJob.Status = job.StatusDead
		currentJob.FinishedAt = &now
	}

	return cloneJob(*currentJob), nil
}

func (q *Queue) Dead() []job.Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	jobs := []job.Job{}

	for _, jobID := range q.order {
		currentJob, ok := q.jobs[jobID]
		if !ok {
			continue
		}

		if currentJob.Status == job.StatusDead {
			jobs = append(jobs, cloneJob(*currentJob))
		}
	}

	return jobs
}

func (q *Queue) requeueExpiredProcessing() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	counter := 0

	for _, jobID := range q.order {
		currentJob, ok := q.jobs[jobID]
		if !ok {
			continue
		}

		if currentJob.Status != job.StatusProcessing {
			continue
		}

		if currentJob.StartedAt == nil {
			continue
		}

		if now.Sub(*currentJob.StartedAt) < q.cfg.VisibilityTimeout {
			continue
		}

		counter++

		currentJob.StartedAt = nil

		if currentJob.Attempts >= currentJob.MaxAttempts {
			currentJob.Status = job.StatusDead
			currentJob.FinishedAt = &now
		} else {
			currentJob.Status = job.StatusQueued
			currentJob.AvailableAt = now
		}
	}

	return counter
}

func (q *Queue) Start(ctx context.Context) {
	ticker := time.NewTicker(q.cfg.SweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.requeueExpiredProcessing()
		}
	}
}
