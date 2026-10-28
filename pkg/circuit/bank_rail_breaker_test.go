package circuit

import "testing"

func TestSelectAvailableRail_FailoverOnOpen(t *testing.T) {
	fednow := &BankRailBreaker{RailName: "FEDNOW_RTP", State: RailClosed, FailureThreshold: 2}
	ach := &BankRailBreaker{RailName: "SAME_DAY_ACH", State: RailClosed, FailureThreshold: 2}

	fednow.RecordFailure()
	fednow.RecordFailure()

	selected := SelectAvailableRail(fednow, ach)
	if selected != "SAME_DAY_ACH" {
		t.Fatalf("expected failover to SAME_DAY_ACH, got %s", selected)
	}
}
