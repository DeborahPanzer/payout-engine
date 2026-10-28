package ach

import "testing"

func TestFormatNachaEntryDetailCCD_Valid94Bytes(t *testing.T) {
	rec, err := FormatNachaEntryDetailCCD("021000021", "987654321", "ACME CORP LLC", "021000020000001", 125000)
	if err != nil {
		t.Fatalf("expected valid NACHA record, got %v", err)
	}
	if len(rec) != 94 {
		t.Fatalf("expected exact 94 characters, got %d", len(rec))
	}
}
