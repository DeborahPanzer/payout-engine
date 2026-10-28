package queue

import "math"

type JobState string

const (
	JobPending    JobState = "PENDING_RETRY"
	JobDeadLetter JobState = "DEAD_LETTER_ISOLATED"
)

type PayoutJob struct {
	PayoutID    string
	Attempt     int
	MaxAttempts int
	BackoffMs   int64
	State       JobState
}

func ScheduleNextRetry(job PayoutJob, baseDelayMs int64, maxDelayMs int64) PayoutJob {
	nextAttempt := job.Attempt + 1
	if nextAttempt >= job.MaxAttempts {
		job.Attempt = nextAttempt
		job.State = JobDeadLetter
		job.BackoffMs = 0
		return job
	}

	exp := float64(baseDelayMs) * math.Pow(2, float64(nextAttempt-1))
	delay := int64(exp)
	if delay > maxDelayMs {
		delay = maxDelayMs
	}

	job.Attempt = nextAttempt
	job.State = JobPending
	job.BackoffMs = delay
	return job
}
