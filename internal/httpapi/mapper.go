package httpapi

import "github.com/AlexandrKudryavtsev/go-job-queue/internal/job"

func toJobResponse(j job.Job) jobResponse {
	return jobResponse{
		ID:          j.ID,
		Type:        j.Type,
		Status:      j.Status,
		Payload:     j.Payload,
		Attempts:    j.Attempts,
		MaxAttempts: j.MaxAttempts,
		LastError:   j.LastError,
		CreatedAt:   j.CreatedAt,
		AvailableAt: j.AvailableAt,
		StartedAt:   j.StartedAt,
		FinishedAt:  j.FinishedAt,
	}
}

func toJobSliceResponse(jobs []job.Job) []jobResponse {
	result := make([]jobResponse, len(jobs))

	for i, currentJob := range jobs {
		result[i] = toJobResponse(currentJob)
	}

	return result
}
