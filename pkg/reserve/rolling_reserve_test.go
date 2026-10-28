package reserve

import (
	"testing"
	"time"
)

func TestCalculateRollingReserve_TenPercent90Days(t *testing.T) {
	now := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	alloc, err := CalculateRollingReserve(100000, ReserveSchedule{ReserveBps: 1000, HoldingWindowDay: 90}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alloc.HeldReserveMinor != 10000 || alloc.ImmediateNet != 90000 {
		t.Fatalf("unexpected allocation: %+v", alloc)
	}
}
