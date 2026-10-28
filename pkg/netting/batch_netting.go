package netting

type GrossObligation struct {
	MerchantID  string
	AmountMinor int64 // Positive for credit payout, negative for debit clawback/fee
}

func CompressMultilateralPositions(items []GrossObligation) map[string]int64 {
	netPositions := make(map[string]int64)
	for _, item := range items {
		netPositions[item.MerchantID] += item.AmountMinor
	}
	for k, v := range netPositions {
		if v <= 0 {
			delete(netPositions, k)
		}
	}
	return netPositions
}
