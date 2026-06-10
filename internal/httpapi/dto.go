package httpapi

import (
	"encoding/json"
	"time"

	"github.com/AlexandrKudryavtsev/go-job-queue/internal/job"
)

type errorResponse struct {
	Error string `json:"error"`
}

type jobResponse struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Status      job.Status      `json:"status"`
	Payload     json.RawMessage `json:"payload"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"maxAttempts"`
	LastError   string          `json:"lastError,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	AvailableAt time.Time       `json:"availableAt"`
	StartedAt   *time.Time      `json:"startedAt,omitempty"`
	FinishedAt  *time.Time      `json:"finishedAt,omitempty"`
}

type createJobRequest struct {
	Type         string          `json:"type"`
	Payload      json.RawMessage `json:"payload"`
	MaxAttempts  int             `json:"maxAttempts"`
	DelaySeconds int             `json:"delaySeconds"`
}
