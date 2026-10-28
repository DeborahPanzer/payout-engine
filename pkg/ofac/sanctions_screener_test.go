package ofac

import "testing"

func TestScreenBeneficiary_BlocksSanctionedEntity(t *testing.T) {
	blocklist := []string{"LAZARUS CYBER GROUP", "DARKSIDE HOLDINGS"}
	res := ScreenBeneficiary("Lazarus Cyber Group, LLC.", blocklist)
	if res.Allowed || !res.ComplianceHold {
		t.Fatalf("expected sanctioned beneficiary to be blocked, got %+v", res)
	}
}
