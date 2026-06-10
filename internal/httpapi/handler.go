package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/AlexandrKudryavtsev/go-job-queue/internal/queue"
)

type Handler struct {
	queue *queue.Queue
}

func New(queue *queue.Queue) *Handler {
	return &Handler{
		queue: queue,
	}
}

func (h *Handler) Register(mx *http.ServeMux) {
	mx.HandleFunc("POST /jobs", h.createJob)
	// mx.HandleFunc("GET /jobs/next", nil)
	// mx.HandleFunc("POST /jobs/{id}/ack", nil)
	// mx.HandleFunc("POST /jobs/{id}/nack", nil)
	// mx.HandleFunc("GET /jobs/dead", nil)
	// mx.HandleFunc("GET /stats", nil)
}

func (h *Handler) createJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	createdJob, err := h.queue.Create(queue.CreateJobInput{
		Type:         req.Type,
		Payload:      req.Payload,
		MaxAttempts:  req.MaxAttempts,
		DelaySeconds: req.DelaySeconds,
	})
	if err != nil {
		writeQueueError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toJobResponse(createdJob))
}
