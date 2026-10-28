package circuit

type RailState string

const (
	RailClosed   RailState = "CLOSED"
	RailOpen     RailState = "OPEN"
	RailHalfOpen RailState = "HALF_OPEN"
)

type BankRailBreaker struct {
	RailName            string
	State               RailState
	ConsecutiveFailures int
	FailureThreshold    int
}

func (b *BankRailBreaker) RecordFailure() {
	b.ConsecutiveFailures++
	if b.ConsecutiveFailures >= b.FailureThreshold {
		b.State = RailOpen
	}
}

func (b *BankRailBreaker) RecordSuccess() {
	b.ConsecutiveFailures = 0
	b.State = RailClosed
}

func SelectAvailableRail(primary, fallback *BankRailBreaker) string {
	if primary.State != RailOpen {
		return primary.RailName
	}
	return fallback.RailName
}
