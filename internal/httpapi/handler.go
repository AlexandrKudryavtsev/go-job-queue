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
	mx.HandleFunc("GET /jobs/next", h.next)
	mx.HandleFunc("GET /jobs/{id}", h.getJob)
	mx.HandleFunc("POST /jobs/{id}/ack", h.ack)
	mx.HandleFunc("POST /jobs/{id}/nack", h.nack)
	mx.HandleFunc("GET /jobs/dead", h.dead)
	mx.HandleFunc("GET /stats", h.stats)
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

func (h *Handler) stats(w http.ResponseWriter, _ *http.Request) {
	stats := h.queue.Stats()

	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) next(w http.ResponseWriter, _ *http.Request) {
	nextJob, err := h.queue.Next()

	if err != nil {
		writeQueueError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toJobResponse(nextJob))
}

func (h *Handler) ack(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ackedJob, err := h.queue.Ack(id)
	if err != nil {
		writeQueueError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toJobResponse(ackedJob))
}

func (h *Handler) nack(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req nackJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	nackedJob, err := h.queue.Nack(id, req.Error)
	if err != nil {
		writeQueueError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toJobResponse(nackedJob))
}

func (h *Handler) dead(w http.ResponseWriter, _ *http.Request) {
	jobs := h.queue.Dead()

	writeJSON(w, http.StatusOK, toJobSliceResponse(jobs))
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	j, err := h.queue.Get(jobID)
	if err != nil {
		writeQueueError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toJobResponse(j))
}
