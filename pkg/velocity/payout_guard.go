package velocity

import "time"

type PayoutSecurityContext struct {
	AmountMinor           int64
	DailyCumulativeMinor  int64
	DailyLimitMinor       int64
	BankAccountUpdatedAt  time.Time
	NowUTC                time.Time
	MandatoryCoolingHours int
}

func EvaluatePayoutVelocity(ctx PayoutSecurityContext) (bool, string) {
	coolingDuration := time.Duration(ctx.MandatoryCoolingHours) * time.Hour
	if ctx.NowUTC.Sub(ctx.BankAccountUpdatedAt) < coolingDuration {
		return false, "BANK_CHANGE_COOLING_PERIOD_ACTIVE"
	}
	if ctx.DailyCumulativeMinor+ctx.AmountMinor > ctx.DailyLimitMinor {
		return false, "DAILY_DISBURSEMENT_VELOCITY_EXCEEDED"
	}
	return true, "ALLOWED"
}
