package split

import "errors"

type SplitRecipient struct {
	AccountID string
	ShareBps  int64
}

type SplitResult struct {
	PlatformFeeMinor int64
	Payouts          map[string]int64
}

func ComputeMarketplaceSplit(grossMinor int64, platformFeeBps int64, recipients []SplitRecipient) (SplitResult, error) {
	if grossMinor <= 0 {
		return SplitResult{}, errors.New("split: gross amount must be positive")
	}
	var totalBps int64 = platformFeeBps
	for _, r := range recipients {
		if r.ShareBps <= 0 {
			return SplitResult{}, errors.New("split: recipient share must be positive")
		}
		totalBps += r.ShareBps
	}
	if totalBps != 10000 {
		return SplitResult{}, errors.New("split: total basis points across platform and recipients must equal 10000")
	}

	platformFee := (grossMinor * platformFeeBps) / 10000
	remaining := grossMinor - platformFee
	payouts := make(map[string]int64, len(recipients))

	for i, r := range recipients {
		if i == len(recipients)-1 {
			payouts[r.AccountID] = remaining
		} else {
			share := (grossMinor * r.ShareBps) / 10000
			payouts[r.AccountID] = share
			remaining -= share
		}
	}

	return SplitResult{
		PlatformFeeMinor: platformFee,
		Payouts:          payouts,
	}, nil
}
