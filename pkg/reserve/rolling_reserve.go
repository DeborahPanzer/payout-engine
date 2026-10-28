package reserve

import (
	"errors"
	"time"
)

type ReserveSchedule struct {
	ReserveBps       int64 // e.g., 1000 bps = 10.00%
	HoldingWindowDay int
}

type ReserveAllocation struct {
	GrossPayoutMinor int64
	HeldReserveMinor int64
	ImmediateNet     int64
	ReleaseAtUTC     time.Time
}

func CalculateRollingReserve(grossMinor int64, sched ReserveSchedule, settledAt time.Time) (ReserveAllocation, error) {
	if grossMinor <= 0 {
		return ReserveAllocation{}, errors.New("reserve: gross payout must be positive")
	}
	if sched.ReserveBps < 0 || sched.ReserveBps > 10000 {
		return ReserveAllocation{}, errors.New("reserve: basis points must be between 0 and 10000")
	}

	held := (grossMinor*sched.ReserveBps + 5000) / 10000
	net := grossMinor - held
	releaseAt := settledAt.UTC().AddDate(0, 0, sched.HoldingWindowDay)

	return ReserveAllocation{
		GrossPayoutMinor: grossMinor,
		HeldReserveMinor: held,
		ImmediateNet:     net,
		ReleaseAtUTC:     releaseAt,
	}, nil
}
