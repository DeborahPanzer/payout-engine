package rtp

import (
	"strings"
	"testing"
)

func TestBuildFedNowPacs008XML(t *testing.T) {
	xml, err := BuildFedNowPacs008XML(Pacs008Transfer{
		MsgID:          "MSG-2026-001",
		EndToEndID:     "E2E-991",
		AmountCents:    500025,
		Currency:       "USD",
		DebtorABA:      "021000021",
		CreditorABA:    "121000248",
		CreditorName:   "MERCHANT ESCROW",
		SettlementDate: "2026-10-28",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(xml, `<IntrBkSttlmAmt Ccy="USD">5000.25</IntrBkSttlmAmt>`) {
		t.Fatalf("missing expected settlement amount tag in xml: %s", xml)
	}
}
