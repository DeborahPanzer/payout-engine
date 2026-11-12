package clawback

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrInvalidClawbackAmount indicates a non-positive clawback obligation or negative balance state.
	ErrInvalidClawbackAmount = errors.New("clawback: obligation must be positive and merchant balances non-negative")
	// ErrClawbackIdempotencyConflict indicates the same ClawbackID was re-submitted with altered parameters.
	ErrClawbackIdempotencyConflict = errors.New("clawback: idempotency conflict on mutated clawback amount or merchant")
)

// MerchantPoolBalances represents the three merchant sub-balances managed before bank disbursement.
type MerchantPoolBalances struct {
	MerchantID          string
	Currency            string
	MaturedReserveMinor int64
	PendingPayableMinor int64
	LockedReserveMinor  int64
}

// ClawbackDirective represents an upstream refund or chargeback deduction request from refund-guard.
type ClawbackDirective struct {
	ClawbackID               string
	MerchantID               string
	Currency                 string
	ObligationMinor          int64
	AllowLockedReserveTap    bool
}

// ClawbackAllocation records the exact zero-sum waterfall offset across merchant pools.
type ClawbackAllocation struct {
	ClawbackID                 string
	MerchantID                 string
	Currency                   string
	RequestedMinor             int64
	DeductedMaturedReserve     int64
	DeductedPendingPayable     int64
	DeductedLockedReserve      int64
	ResidualDebitRecoveryMinor int64
	UpdatedBalances            MerchantPoolBalances
}

// WaterfallEngine executes idempotent, conservation-safe merchant clawbacks.
type WaterfallEngine struct {
	mu       sync.Mutex
	balances map[string]MerchantPoolBalances
	receipts map[string]ClawbackAllocation
}

// NewWaterfallEngine initializes an in-memory merchant clawback waterfall coordinator.
func NewWaterfallEngine() *WaterfallEngine {
	return &WaterfallEngine{
		balances: make(map[string]MerchantPoolBalances),
		receipts: make(map[string]ClawbackAllocation),
	}
}

func poolKey(merchantID, currency string) string {
	return merchantID + ":" + currency
}

// SeedMerchantPools registers or updates non-negative merchant liquidity pools.
func (e *WaterfallEngine) SeedMerchantPools(state MerchantPoolBalances) error {
	if state.MerchantID == "" || state.Currency == "" ||
		state.MaturedReserveMinor < 0 || state.PendingPayableMinor < 0 || state.LockedReserveMinor < 0 {
		return ErrInvalidClawbackAmount
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.balances[poolKey(state.MerchantID, state.Currency)] = state
	return nil
}

// ApplyClawback deducts obligations in strict priority:
// 1. Matured Rolling Reserve -> 2. Pending Intraday Payable -> 3. Locked Chargeback Reserve (if authorized) -> 4. Residual ACH Debit Recovery.
func (e *WaterfallEngine) ApplyClawback(dir ClawbackDirective) (ClawbackAllocation, error) {
	if dir.ClawbackID == "" || dir.MerchantID == "" || dir.Currency == "" || dir.ObligationMinor <= 0 {
		return ClawbackAllocation{}, ErrInvalidClawbackAmount
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if cached, exists := e.receipts[dir.ClawbackID]; exists {
		if cached.MerchantID != dir.MerchantID ||
			cached.Currency != dir.Currency ||
			cached.RequestedMinor != dir.ObligationMinor {
			return ClawbackAllocation{}, fmt.Errorf("%w: id=%s", ErrClawbackIdempotencyConflict, dir.ClawbackID)
		}
		return cached, nil
	}

	key := poolKey(dir.MerchantID, dir.Currency)
	bal := e.balances[key]
	bal.MerchantID = dir.MerchantID
	bal.Currency = dir.Currency

	remaining := dir.ObligationMinor

	// Step 1: Offset against Matured Reserve
	fromMatured := minInt64(bal.MaturedReserveMinor, remaining)
	bal.MaturedReserveMinor -= fromMatured
	remaining -= fromMatured

	// Step 2: Offset against Pending Intraday Payable (reduces next wire/ACH payout)
	fromPayable := minInt64(bal.PendingPayableMinor, remaining)
	bal.PendingPayableMinor -= fromPayable
	remaining -= fromPayable

	// Step 3: Offset against Locked Rolling Reserve if card-network dispute tap is authorized
	var fromLocked int64
	if dir.AllowLockedReserveTap && remaining > 0 {
		fromLocked = minInt64(bal.LockedReserveMinor, remaining)
		bal.LockedReserveMinor -= fromLocked
		remaining -= fromLocked
	}

	// Step 4: Any residual becomes an explicit Direct Debit Recovery receivable
	residual := remaining

	e.balances[key] = bal
	alloc := ClawbackAllocation{
		ClawbackID:                 dir.ClawbackID,
		MerchantID:                 dir.MerchantID,
		Currency:                   dir.Currency,
		RequestedMinor:             dir.ObligationMinor,
		DeductedMaturedReserve:     fromMatured,
		DeductedPendingPayable:     fromPayable,
		DeductedLockedReserve:      fromLocked,
		ResidualDebitRecoveryMinor: residual,
		UpdatedBalances:            bal,
	}
	e.receipts[dir.ClawbackID] = alloc
	return alloc, nil
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
