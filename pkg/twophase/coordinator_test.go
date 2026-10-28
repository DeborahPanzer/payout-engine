package twophase

import "testing"

func TestTwoPhaseCommit_UnanimousSuccess(t *testing.T) {
	tx := &BatchDisbursementTx{
		BatchID:    "batch_01",
		TotalMinor: 250000,
		Currency:   "USD",
		State:      StateInitiated,
		Participants: []ParticipantVote{
			{NodeID: "ledger_lock", Prepared: true},
			{NodeID: "ofac_guard", Prepared: true},
		},
	}

	if err := EvaluatePreparePhase(tx); err != nil {
		t.Fatalf("expected prepare success, got %v", err)
	}
	if err := CommitBatch(tx); err != nil {
		t.Fatalf("expected commit success, got %v", err)
	}
	if tx.State != StateCommitted {
		t.Fatalf("expected state COMMITTED, got %s", tx.State)
	}
}

func TestTwoPhaseCommit_SingleAbortRollsBack(t *testing.T) {
	tx := &BatchDisbursementTx{
		BatchID:    "batch_02",
		TotalMinor: 250000,
		Currency:   "USD",
		State:      StateInitiated,
		Participants: []ParticipantVote{
			{NodeID: "ledger_lock", Prepared: true},
			{NodeID: "bank_rail", Prepared: false, Reason: "rail timeout"},
		},
	}

	if err := EvaluatePreparePhase(tx); err == nil {
		t.Fatal("expected prepare phase to fail on dissenting participant")
	}
	if tx.State != StateAborted {
		t.Fatalf("expected state ABORTED, got %s", tx.State)
	}
}
