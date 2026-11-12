package noc

import (
	"errors"
	"testing"
	"time"
)

func TestParseAddenda98_ValidAndMalformedCases(t *testing.T) {
	received := time.Date(2026, time.November, 12, 15, 0, 0, 0, time.UTC) // Thursday

	// 1. C03 combined ABA Routing (021000021 - JPMorgan Chase valid Mod-10) + Account
	c03, err := ParseAddenda98("C03", "021000021   98765432101", "021000020000123", received)
	if err != nil {
		t.Fatalf("expected valid C03 parse, got error: %v", err)
	}
	if !c03.RequiresRoutingUpdate || !c03.RequiresAccountUpdate {
		t.Fatalf("expected routing and account update flags on C03, got %+v", c03)
	}
	if c03.NewRoutingNumber != "021000021" || c03.NewAccountNumber != "98765432101" {
		t.Fatalf("unexpected C03 extracted values: %+v", c03)
	}
	// 6 banking days from Thursday Nov 12 skips weekends -> Friday Nov 20
	if got := c03.ComplianceDeadlineUTC.Format("2006-01-02"); got != "2026-11-20" {
		t.Fatalf("expected 6-banking-day SLA deadline 2026-11-20, got %s", got)
	}

	// 2. C02 with corrupted ABA Mod-10 checksum -> Must reject before corrupting beneficiary vault
	_, badRoutingErr := ParseAddenda98("C02", "021000029", "021000020000124", received)
	if !errors.Is(badRoutingErr, ErrInvalidABARoutingChecksum) {
		t.Fatalf("expected ErrInvalidABARoutingChecksum, got: %v", badRoutingErr)
	}

	// 3. C07 full triple update (Routing + Account + Savings Credit 32)
	c07, err := ParseAddenda98("C07", "121000358 4455667788 32", "121000350000999", received)
	if err != nil {
		t.Fatalf("expected valid C07 parse, got: %v", err)
	}
	if c07.NewTransactionCode != "32" || !c07.RequiresTranCodeUpdate {
		t.Fatalf("expected transaction code 32 update, got %+v", c07)
	}

	// 4. Unsupported change code -> Reject cleanly
	_, unknownErr := ParseAddenda98("C99", "021000021", "021000020000125", received)
	if !errors.Is(unknownErr, ErrUnsupportedChangeCode) {
		t.Fatalf("expected ErrUnsupportedChangeCode, got: %v", unknownErr)
	}
}
