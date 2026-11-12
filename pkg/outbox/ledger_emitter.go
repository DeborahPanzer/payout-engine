package outbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var (
	// ErrWeakSigningSecret indicates an HMAC secret shorter than 32 characters.
	ErrWeakSigningSecret = errors.New("outbox: HMAC signing secret must be at least 32 characters")
	// ErrInvalidOutboxPosting indicates invalid currency, non-positive amount, or self-referencing ledger accounts.
	ErrInvalidOutboxPosting = errors.New("outbox: invalid posting parameters or identical debit/credit accounts")
	// ErrOutboxIdempotencyConflict indicates an idempotency key was reused with a different canonical digest.
	ErrOutboxIdempotencyConflict = errors.New("outbox: idempotency key reused with mutated financial payload")
)

// FinancialEventEnvelope matches the ecosystem-wide v1.2.0 wire contract shared with checkout-core and ledger-vault.
type FinancialEventEnvelope struct {
	EventID           string
	SchemaVersion     string
	EventType         string
	SourceService     string
	TargetService     string
	IdempotencyKey    string
	DisbursementID    string
	CorrelationID     string
	CausationID       string
	OccurredAtEpochMs int64
	Currency          string
	AmountMinorStr    string
	DebitAccountCode  string
	CreditAccountCode string
	PayloadSHA256     string
	HMACSignature     string
}

// EnqueueInput defines the parameters for emitting a signed treasury settlement or reversal event.
type EnqueueInput struct {
	EventID           string
	EventType         string
	TargetService     string
	IdempotencyKey    string
	DisbursementID    string
	CorrelationID     string
	CausationID       string
	OccurredAtEpochMs int64
	Currency          string
	AmountMinor       int64
	DebitAccountCode  string
	CreditAccountCode string
}

type outboxRecord struct {
	envelope   FinancialEventEnvelope
	dispatched bool
}

// TransactionalOutbox manages replay-safe, HMAC-SHA256 authenticated ledger postings.
type TransactionalOutbox struct {
	mu            sync.Mutex
	signingSecret []byte
	orderedIDs    []string
	byEventID     map[string]*outboxRecord
	byIdemKey     map[string]*outboxRecord
}

// NewTransactionalOutbox creates a cryptographic outbox emitter for payout-engine.
func NewTransactionalOutbox(secret string) (*TransactionalOutbox, error) {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < 32 {
		return nil, ErrWeakSigningSecret
	}
	return &TransactionalOutbox{
		signingSecret: []byte(trimmed),
		byEventID:     make(map[string]*outboxRecord),
		byIdemKey:     make(map[string]*outboxRecord),
	}, nil
}

func computeCanonicalDigest(env FinancialEventEnvelope) string {
	canonical := strings.Join([]string{
		"v1.2.0",
		"payout-engine",
		env.TargetService,
		env.EventID,
		env.EventType,
		env.IdempotencyKey,
		env.DisbursementID,
		env.CorrelationID,
		env.CausationID,
		env.Currency,
		env.AmountMinorStr,
		env.DebitAccountCode,
		env.CreditAccountCode,
	}, "|")

	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func computeHMAC(secret []byte, occurredAtEpochMs int64, payloadSHA256 string) string {
	mac := hmac.New(sha256.New, secret)
	signingInput := strconv.FormatInt(occurredAtEpochMs, 10) + "." + payloadSHA256
	_, _ = mac.Write([]byte(signingInput))
	return hex.EncodeToString(mac.Sum(nil))
}

// Enqueue validates zero-sum double-entry legs, computes the SHA-256 & HMAC signatures, and stores the envelope.
func (o *TransactionalOutbox) Enqueue(in EnqueueInput) (FinancialEventEnvelope, error) {
	curr := strings.ToUpper(strings.TrimSpace(in.Currency))
	debit := strings.TrimSpace(in.DebitAccountCode)
	credit := strings.TrimSpace(in.CreditAccountCode)

	if in.AmountMinor <= 0 || len(curr) != 3 || debit == "" || credit == "" || debit == credit ||
		in.EventID == "" || in.IdempotencyKey == "" {
		return FinancialEventEnvelope{}, ErrInvalidOutboxPosting
	}

	env := FinancialEventEnvelope{
		EventID:           in.EventID,
		SchemaVersion:     "1.2.0",
		EventType:         in.EventType,
		SourceService:     "payout-engine",
		TargetService:     in.TargetService,
		IdempotencyKey:    in.IdempotencyKey,
		DisbursementID:    in.DisbursementID,
		CorrelationID:     in.CorrelationID,
		CausationID:       in.CausationID,
		OccurredAtEpochMs: in.OccurredAtEpochMs,
		Currency:          curr,
		AmountMinorStr:    strconv.FormatInt(in.AmountMinor, 10),
		DebitAccountCode:  debit,
		CreditAccountCode: credit,
	}
	env.PayloadSHA256 = computeCanonicalDigest(env)

	o.mu.Lock()
	defer o.mu.Unlock()

	if existing, ok := o.byIdemKey[in.IdempotencyKey]; ok {
		if existing.envelope.PayloadSHA256 != env.PayloadSHA256 {
			return FinancialEventEnvelope{}, fmt.Errorf("%w: key=%s", ErrOutboxIdempotencyConflict, in.IdempotencyKey)
		}
		return existing.envelope, nil
	}

	env.HMACSignature = computeHMAC(o.signingSecret, env.OccurredAtEpochMs, env.PayloadSHA256)
	rec := &outboxRecord{envelope: env, dispatched: false}
	o.orderedIDs = append(o.orderedIDs, env.EventID)
	o.byEventID[env.EventID] = rec
	o.byIdemKey[env.IdempotencyKey] = rec
	return env, nil
}

// DrainPending returns up to limit undispatched envelopes in strict FIFO causal order.
func (o *TransactionalOutbox) DrainPending(limit int) []FinancialEventEnvelope {
	if limit <= 0 {
		return nil
	}
	o.mu.Lock()
	defer o.mu.Unlock()

	batch := make([]FinancialEventEnvelope, 0, limit)
	for _, id := range o.orderedIDs {
		rec := o.byEventID[id]
		if rec != nil && !rec.dispatched {
			batch = append(batch, rec.envelope)
			if len(batch) >= limit {
				break
			}
		}
	}
	return batch
}

// Acknowledge marks an envelope as delivered to ledger-vault.
func (o *TransactionalOutbox) Acknowledge(eventID string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	rec, ok := o.byEventID[eventID]
	if !ok {
		return false
	}
	rec.dispatched = true
	return true
}

// VerifyEnvelopeIntegrity performs constant-time HMAC-SHA256 and replay-window verification.
func VerifyEnvelopeIntegrity(env FinancialEventEnvelope, secret string, nowEpochMs, toleranceMs int64) bool {
	delta := nowEpochMs - env.OccurredAtEpochMs
	if delta < 0 {
		delta = -delta
	}
	if delta > toleranceMs {
		return false
	}

	expectedDigest := computeCanonicalDigest(env)
	if expectedDigest != env.PayloadSHA256 {
		return false
	}

	expectedHex := computeHMAC([]byte(strings.TrimSpace(secret)), env.OccurredAtEpochMs, expectedDigest)
	expectedMAC, err1 := hex.DecodeString(expectedHex)
	providedMAC, err2 := hex.DecodeString(env.HMACSignature)
	if err1 != nil || err2 != nil || len(expectedMAC) == 0 {
		return false
	}
	return hmac.Equal(expectedMAC, providedMAC)
}
