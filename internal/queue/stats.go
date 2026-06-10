package queue

import (
	"github.com/AlexandrKudryavtsev/go-job-queue/internal/job"
)

type Stats struct {
	Queued     int `json:"queued"`
	Processing int `json:"processing"`
	Done       int `json:"done"`
	Dead       int `json:"dead"`
	Delayed    int `json:"delayed"`
}

func (q *Queue) Stats() Stats {
	var stats Stats

	q.mu.Lock()
	defer q.mu.Unlock()

	for _, jobID := range q.order {
		currentJob, ok := q.jobs[jobID]
		if !ok {
			continue
		}

		switch currentJob.Status {
		case job.StatusQueued:
			stats.Queued++
		case job.StatusProcessing:
			stats.Processing++
		case job.StatusDone:
			stats.Done++
		case job.StatusDead:
			stats.Dead++
		case job.StatusDelayed:
			stats.Delayed++
		}
	}

	return stats
}
