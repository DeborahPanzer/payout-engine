package velocity

import (
	"testing"
	"time"
)

func TestEvaluatePayoutVelocity_CoolingPeriodBlocksPayout(t *testing.T) {
	now := time.Now().UTC()
	ok, code := EvaluatePayoutVelocity(PayoutSecurityContext{
		AmountMinor:           10000,
		DailyCumulativeMinor:  0,
		DailyLimitMinor:       500000,
		BankAccountUpdatedAt:  now.Add(-2 * time.Hour),
		NowUTC:                now,
		MandatoryCoolingHours: 24,
	})
	if ok || code != "BANK_CHANGE_COOLING_PERIOD_ACTIVE" {
		t.Fatalf("expected cooling period block, got ok=%v code=%s", ok, code)
	}
}
