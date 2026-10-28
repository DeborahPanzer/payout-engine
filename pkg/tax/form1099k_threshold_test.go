package tax

import "testing"

func TestEvaluateIRS1099KWithholding_UnverifiedTIN(t *testing.T) {
	res := EvaluateIRS1099KWithholding(MerchantTaxProfile{
		TINVerified:        false,
		YTDGrossMinor:      55000,
		CurrentPayoutMinor: 10000,
		ReportingThreshold: 60000,
	})
	if !res.Requires1099K || res.BackupWithholdingMinor != 2400 {
		t.Fatalf("expected 24%% backup withholding (2400 cents), got %+v", res)
	}
}
