package queue

import "testing"

func TestScheduleNextRetry_DeadLetterPromotion(t *testing.T) {
	job := PayoutJob{PayoutID: "po_1", Attempt: 4, MaxAttempts: 5}
	updated := ScheduleNextRetry(job, 1000, 60000)
	if updated.State != JobDeadLetter {
		t.Fatalf("expected DEAD_LETTER_ISOLATED, got %s", updated.State)
	}
}
