package outbox

import (
	"errors"
	"testing"
)

const testSharedSecret = "9f8e7d6c5b4a39281706f5e4d3c2b1a09f8e7d6c5b4a39281706f5e4d3c2b1a0"

func TestTransactionalOutbox_SigningReplayAndIdempotency(t *testing.T) {
	ob, err := NewTransactionalOutbox(testSharedSecret)
	if err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}

	nowMs := int64(1_762_940_000_000)
	input := EnqueueInput{
		EventID:           "evt_po_settled_01",
		EventType:         "payout.rail.settled",
		TargetService:     "ledger-vault",
		IdempotencyKey:    "idem_po_batch_8801",
		DisbursementID:    "po_batch_8801",
		CorrelationID:     "corr_8801",
		CausationID:       "pacs002_accc_8801",
		OccurredAtEpochMs: nowMs,
		Currency:          "usd",
		AmountMinor:       425_000,
		DebitAccountCode:  "LIABILITY:MERCHANT_PAYABLE_USD",
		CreditAccountCode: "ASSET:NOSTRO_FEDNOW_USD",
	}

	env, err := ob.Enqueue(input)
	if err != nil {
		t.Fatalf("expected clean enqueue, got: %v", err)
	}

	if !VerifyEnvelopeIntegrity(env, testSharedSecret, nowMs+10_000, 300_000) {
		t.Fatal("expected valid HMAC verification within 300s window")
	}

	// Tampered amount must fail verification
	tampered := env
	tampered.AmountMinorStr = "999999"
	if VerifyEnvelopeIntegrity(tampered, testSharedSecret, nowMs+10_000, 300_000) {
		t.Fatal("expected tampered amount to fail digest/HMAC verification")
	}

	// Expired replay window (> 300s) must fail verification
	if VerifyEnvelopeIntegrity(env, testSharedSecret, nowMs+305_000, 300_000) {
		t.Fatal("expected replay outside 300s window to be rejected")
	}

	// Mutated payload on existing IdempotencyKey must raise ErrOutboxIdempotencyConflict
	mutatedInput := input
	mutatedInput.AmountMinor = 500_000
	if _, err := ob.Enqueue(mutatedInput); !errors.Is(err, ErrOutboxIdempotencyConflict) {
		t.Fatalf("expected ErrOutboxIdempotencyConflict, got: %v", err)
	}

	// FIFO drain and acknowledgement
	if len(ob.DrainPending(10)) != 1 {
		t.Fatal("expected 1 pending envelope before acknowledgement")
	}
	if !ob.Acknowledge("evt_po_settled_01") || len(ob.DrainPending(10)) != 0 {
		t.Fatal("expected 0 pending envelopes after acknowledgement")
	}
}
