package merkle

import "testing"

func TestComputePayoutBatchMerkleRoot(t *testing.T) {
	root1, err := ComputePayoutBatchMerkleRoot([]string{"po_101", "po_102", "po_103"})
	if err != nil || len(root1) != 64 {
		t.Fatalf("expected 64-char hex root, got %s (err=%v)", root1, err)
	}
	root2, _ := ComputePayoutBatchMerkleRoot([]string{"po_101", "po_102", "po_999"})
	if root1 == root2 {
		t.Fatal("expected mutated payout leaf to produce distinct Merkle root")
	}
}
