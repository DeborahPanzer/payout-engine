package netting

import "testing"

func TestCompressMultilateralPositions(t *testing.T) {
	net := CompressMultilateralPositions([]GrossObligation{
		{MerchantID: "m_1", AmountMinor: 50000},
		{MerchantID: "m_1", AmountMinor: -12000},
		{MerchantID: "m_2", AmountMinor: -5000},
	})
	if net["m_1"] != 38000 {
		t.Fatalf("expected net 38000 for m_1, got %d", net["m_1"])
	}
	if _, exists := net["m_2"]; exists {
		t.Fatal("expected negative net position m_2 to be withheld from outbound wire batch")
	}
}
