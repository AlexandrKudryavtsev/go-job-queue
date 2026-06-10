package queue

import "errors"

var (
	ErrJobNotFound        = errors.New("job not found")
	ErrInvalidStatus      = errors.New("invalid job status")
	ErrNoJobsAvailable    = errors.New("no jobs available")
	ErrInvalidJobType     = errors.New("invalid job type")
	ErrInvalidPayload     = errors.New("invalid job payload")
	ErrInvalidMaxAttempts = errors.New("invalid job max attempts")
	ErrInvalidDelay       = errors.New("invalid job delay")
)
