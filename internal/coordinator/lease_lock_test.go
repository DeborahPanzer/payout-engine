package coordinator

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLeaseCoordinator_ConcurrentContentionAndFencingTakeover(t *testing.T) {
	coord := NewLeaseCoordinator()
	baseNow := time.Date(2026, time.November, 12, 14, 0, 0, 0, time.UTC)
	ttl := 30 * time.Second

	const workers = 200
	var wg sync.WaitGroup
	var winners atomic.Int32
	var winningToken LeaseToken
	var tokenMu sync.Mutex

	for i := 0; i < workers; i++ {
		wg.Add(1)
		workerID := fmt.Sprintf("worker-%03d", i)
		go func(id string) {
			defer wg.Done()
			tok, err := coord.Acquire("batch_ach_20261112", id, ttl, baseNow)
			if err == nil {
				winners.Add(1)
				tokenMu.Lock()
				winningToken = tok
				tokenMu.Unlock()
			} else if !errors.Is(err, ErrBatchLeaseHeld) {
				t.Errorf("unexpected acquire error: %v", err)
			}
		}(workerID)
	}
	wg.Wait()

	if winners.Load() != 1 {
		t.Fatalf("expected exactly 1 lease winner among %d goroutines, got %d", workers, winners.Load())
	}

	if err := coord.ValidateFencingToken("batch_ach_20261112", winningToken.FencingEpoch, baseNow.Add(10*time.Second)); err != nil {
		t.Fatalf("expected active fencing token to validate, got: %v", err)
	}

	// Advance clock past TTL so a recovery worker takes over the stalled batch.
	afterExpiry := baseNow.Add(31 * time.Second)
	takeoverToken, err := coord.Acquire("batch_ach_20261112", "worker-recovery", ttl, afterExpiry)
	if err != nil {
		t.Fatalf("expected takeover after TTL expiry to succeed, got: %v", err)
	}
	if takeoverToken.FencingEpoch <= winningToken.FencingEpoch {
		t.Fatalf("expected strictly increasing fencing epoch (%d <= %d)", takeoverToken.FencingEpoch, winningToken.FencingEpoch)
	}

	// Zombie worker wakes up and attempts to validate its stale epoch: must be blocked.
	zombieErr := coord.ValidateFencingToken("batch_ach_20261112", winningToken.FencingEpoch, afterExpiry.Add(time.Second))
	if !errors.Is(zombieErr, ErrStaleFencingToken) {
		t.Fatalf("expected ErrStaleFencingToken for superseded epoch, got: %v", zombieErr)
	}

	if err := coord.Release("batch_ach_20261112", "worker-recovery", takeoverToken.FencingEpoch); err != nil {
		t.Fatalf("expected clean release by takeover owner, got: %v", err)
	}
}
