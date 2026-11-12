package clawback

import (
	"errors"
	"testing"
)

func TestWaterfallEngine_ConservationAndIdempotency(t *testing.T) {
	engine := NewWaterfallEngine()
	err := engine.SeedMerchantPools(MerchantPoolBalances{
		MerchantID:          "acct_merch_01",
		Currency:            "USD",
		MaturedReserveMinor: 30_000, // $300.00
		PendingPayableMinor: 50_000, // $500.00
		LockedReserveMinor:  40_000, // $400.00
	})
	if err != nil {
		t.Fatalf("unexpected seed error: %v", err)
	}

	// Request $1,500.00 chargeback clawback with locked reserve tap authorized
	dir := ClawbackDirective{
		ClawbackID:            "cbk_dispute_7701",
		MerchantID:            "acct_merch_01",
		Currency:              "USD",
		ObligationMinor:       150_000,
		AllowLockedReserveTap: true,
	}

	alloc, err := engine.ApplyClawback(dir)
	if err != nil {
		t.Fatalf("unexpected clawback error: %v", err)
	}

	// Conservation invariant: 30,000 + 50,000 + 40,000 + 30,000 residual == 150,000
	sum := alloc.DeductedMaturedReserve + alloc.DeductedPendingPayable + alloc.DeductedLockedReserve + alloc.ResidualDebitRecoveryMinor
	if sum != dir.ObligationMinor {
		t.Fatalf("conservation invariant violated: sum=%d expected=%d", sum, dir.ObligationMinor)
	}
	if alloc.ResidualDebitRecoveryMinor != 30_000 {
		t.Fatalf("expected $300.00 (30000 minor) residual ACH debit recovery, got %d", alloc.ResidualDebitRecoveryMinor)
	}

	// Idempotent retry returns identical allocation without double-deducting
	retryAlloc, err := engine.ApplyClawback(dir)
	if err != nil || retryAlloc != alloc {
		t.Fatalf("expected idempotent cached receipt, got %+v (err=%v)", retryAlloc, err)
	}

	// Mutated amount on same ClawbackID must trigger ErrClawbackIdempotencyConflict
	_, conflictErr := engine.ApplyClawback(ClawbackDirective{
		ClawbackID:            "cbk_dispute_7701",
		MerchantID:            "acct_merch_01",
		Currency:              "USD",
		ObligationMinor:       160_000,
		AllowLockedReserveTap: true,
	})
	if !errors.Is(conflictErr, ErrClawbackIdempotencyConflict) {
		t.Fatalf("expected ErrClawbackIdempotencyConflict, got: %v", conflictErr)
	}
}
