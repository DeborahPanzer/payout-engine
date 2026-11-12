package coordinator

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrInvalidLeaseInput indicates empty identifiers or non-positive TTL duration.
	ErrInvalidLeaseInput = errors.New("coordinator: invalid batch ID, worker ID, or lease TTL")
	// ErrBatchLeaseHeld indicates another active worker currently holds an unexpired 2PC lease.
	ErrBatchLeaseHeld = errors.New("coordinator: disbursement batch lease is actively held")
	// ErrLeaseNotFound indicates no active lease exists for the target batch.
	ErrLeaseNotFound = errors.New("coordinator: disbursement batch lease not found")
	// ErrStaleFencingToken indicates a zombie worker attempted to commit using a superseded epoch.
	ErrStaleFencingToken = errors.New("coordinator: stale fencing epoch token rejected")
	// ErrLeaseExpired indicates the worker's lease window elapsed before phase-2 commit.
	ErrLeaseExpired = errors.New("coordinator: disbursement batch lease has expired")
)

// LeaseToken represents a monotonically fenced Two-Phase Commit ownership grant.
type LeaseToken struct {
	BatchID      string
	WorkerID     string
	FencingEpoch uint64
	AcquiredAt   time.Time
	ExpiresAt    time.Time
}

// LeaseCoordinator isolates in-memory 2PC batch locks with strictly increasing fencing tokens.
type LeaseCoordinator struct {
	mu          sync.Mutex
	globalEpoch uint64
	activeLocks map[string]LeaseToken
}

// NewLeaseCoordinator initializes an isolated fencing-token lease coordinator.
func NewLeaseCoordinator() *LeaseCoordinator {
	return &LeaseCoordinator{
		activeLocks: make(map[string]LeaseToken),
	}
}

// Acquire grants an exclusive batch lease or renews it for the same worker, incrementing the fencing epoch.
func (c *LeaseCoordinator) Acquire(batchID, workerID string, ttl time.Duration, now time.Time) (LeaseToken, error) {
	if batchID == "" || workerID == "" || ttl <= 0 {
		return LeaseToken{}, ErrInvalidLeaseInput
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.activeLocks[batchID]; ok {
		if now.Before(existing.ExpiresAt) && existing.WorkerID != workerID {
			return LeaseToken{}, fmt.Errorf(
				"%w: batch=%s owner=%s expires=%s",
				ErrBatchLeaseHeld,
				batchID,
				existing.WorkerID,
				existing.ExpiresAt.UTC().Format(time.RFC3339Nano),
			)
		}
	}

	c.globalEpoch++
	grant := LeaseToken{
		BatchID:      batchID,
		WorkerID:     workerID,
		FencingEpoch: c.globalEpoch,
		AcquiredAt:   now.UTC(),
		ExpiresAt:    now.Add(ttl).UTC(),
	}
	c.activeLocks[batchID] = grant
	return grant, nil
}

// ValidateFencingToken verifies that a worker still holds the latest unexpired fencing epoch before wire submission.
func (c *LeaseCoordinator) ValidateFencingToken(batchID string, fencingEpoch uint64, now time.Time) error {
	if batchID == "" {
		return ErrInvalidLeaseInput
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current, ok := c.activeLocks[batchID]
	if !ok {
		return ErrLeaseNotFound
	}
	if fencingEpoch != current.FencingEpoch {
		return fmt.Errorf(
			"%w: provided=%d current=%d",
			ErrStaleFencingToken,
			fencingEpoch,
			current.FencingEpoch,
		)
	}
	if !now.Before(current.ExpiresAt) {
		return ErrLeaseExpired
	}
	return nil
}

// Release explicitly frees an active lease if and only if the caller presents the matching worker ID and epoch.
func (c *LeaseCoordinator) Release(batchID, workerID string, fencingEpoch uint64) error {
	if batchID == "" || workerID == "" {
		return ErrInvalidLeaseInput
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current, ok := c.activeLocks[batchID]
	if !ok {
		return ErrLeaseNotFound
	}
	if current.WorkerID != workerID || current.FencingEpoch != fencingEpoch {
		return ErrStaleFencingToken
	}

	delete(c.activeLocks, batchID)
	return nil
}
