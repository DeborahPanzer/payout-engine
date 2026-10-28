package sepa

import "testing"

func TestValidateIBAN_Mod97(t *testing.T) {
	if !ValidateIBAN("GB82 WEST 1234 5698 7654 32") {
		t.Fatal("expected valid UK IBAN to pass ISO-13616 Mod-97 check")
	}
	if ValidateIBAN("GB00 WEST 1234 5698 7654 32") {
		t.Fatal("expected corrupted IBAN check digits to fail")
	}
}
