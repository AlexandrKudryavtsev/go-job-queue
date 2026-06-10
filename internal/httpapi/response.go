package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AlexandrKudryavtsev/go-job-queue/internal/queue"
)

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeQueueError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, queue.ErrInvalidDelay),
		errors.Is(err, queue.ErrInvalidPayload),
		errors.Is(err, queue.ErrInvalidMaxAttempts),
		errors.Is(err, queue.ErrInvalidJobType):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, queue.ErrJobNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, queue.ErrNoJobsAvailable):
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, queue.ErrInvalidStatus):
		writeError(w, http.StatusConflict, err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
