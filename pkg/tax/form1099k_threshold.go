package tax

type MerchantTaxProfile struct {
	TINVerified        bool
	YTDGrossMinor      int64
	CurrentPayoutMinor int64
	ReportingThreshold int64 // e.g., $600 or $5,000 in cents
}

type TaxWithholdingResult struct {
	Requires1099K          bool
	BackupWithholdingMinor int64
	NetDisbursableMinor    int64
}

func EvaluateIRS1099KWithholding(p MerchantTaxProfile) TaxWithholdingResult {
	newYTD := p.YTDGrossMinor + p.CurrentPayoutMinor
	requiresForm := newYTD >= p.ReportingThreshold

	var withheld int64
	// US statutory backup withholding rate is 24% (2400 bps) if TIN/W-9 is unverified
	if requiresForm && !p.TINVerified {
		withheld = (p.CurrentPayoutMinor * 2400) / 10000
	}

	return TaxWithholdingResult{
		Requires1099K:          requiresForm,
		BackupWithholdingMinor: withheld,
		NetDisbursableMinor:    p.CurrentPayoutMinor - withheld,
	}
}
