package liquidity

import "errors"

type NostroPool struct {
	PoolID           string
	Currency         string
	AvailableMinor   int64
	MinimumBufferMin int64
	PerTxCostMinor   int64
}

func SelectOptimalNostroPool(pools []NostroPool, currency string, payoutMinor int64) (NostroPool, error) {
	var best *NostroPool
	for i := range pools {
		p := &pools[i]
		if p.Currency != currency {
			continue
		}
		if p.AvailableMinor-payoutMinor < p.MinimumBufferMin {
			continue
		}
		if best == nil || p.PerTxCostMinor < best.PerTxCostMinor {
			best = p
		}
	}
	if best == nil {
		return NostroPool{}, errors.New("liquidity: insufficient Nostro pool balance above safety buffer")
	}
	return *best, nil
}
