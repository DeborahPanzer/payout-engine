package returncodes

import "testing"

func TestClassifyACHReturnCode(t *testing.T) {
	r02 := ClassifyACHReturnCode("R02")
	if r02.Retryable || !r02.InvalidateBankAccount {
		t.Fatalf("expected R02 Account Closed to invalidate bank account without retry")
	}
}
