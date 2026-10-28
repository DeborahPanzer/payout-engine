package split

import "testing"

func TestComputeMarketplaceSplit_ExactConservation(t *testing.T) {
	res, err := ComputeMarketplaceSplit(10001, 500, []SplitRecipient{
		{AccountID: "acct_vendor_1", ShareBps: 6000},
		{AccountID: "acct_affiliate_2", ShareBps: 3500},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	total := res.PlatformFeeMinor + res.Payouts["acct_vendor_1"] + res.Payouts["acct_affiliate_2"]
	if total != 10001 {
		t.Fatalf("expected exact sum 10001, got %d", total)
	}
}
