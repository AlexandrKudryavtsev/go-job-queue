package job

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusDelayed    Status = "delayed"
	StatusDead       Status = "dead"
)

type Job struct {
	ID          string
	Type        string
	Status      Status
	Payload     json.RawMessage
	Attempts    int
	MaxAttempts int
	LastError   string
	CreatedAt   time.Time
	AvailableAt time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
}
