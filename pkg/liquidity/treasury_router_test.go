package liquidity

import "testing"

func TestSelectOptimalNostroPool(t *testing.T) {
	pools := []NostroPool{
		{PoolID: "jpm_usd", Currency: "USD", AvailableMinor: 500000, MinimumBufferMin: 100000, PerTxCostMinor: 25},
		{PoolID: "citi_usd", Currency: "USD", AvailableMinor: 500000, MinimumBufferMin: 100000, PerTxCostMinor: 10},
	}
	chosen, err := SelectOptimalNostroPool(pools, "USD", 200000)
	if err != nil || chosen.PoolID != "citi_usd" {
		t.Fatalf("expected citi_usd pool, got %+v (err=%v)", chosen, err)
	}
}
