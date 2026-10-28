package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func sha256Hex(in string) string {
	h := sha256.Sum256([]byte(in))
	return hex.EncodeToString(h[:])
}

func ComputePayoutBatchMerkleRoot(payoutIDs []string) (string, error) {
	if len(payoutIDs) == 0 {
		return "", errors.New("merkle: cannot seal empty payout batch")
	}
	layer := make([]string, len(payoutIDs))
	for i, id := range payoutIDs {
		layer[i] = sha256Hex("PAYOUT_LEAF:" + id)
	}

	for len(layer) > 1 {
		var next []string
		for i := 0; i < len(layer); i += 2 {
			left := layer[i]
			right := left
			if i+1 < len(layer) {
				right = layer[i+1]
			}
			next = append(next, sha256Hex("PAYOUT_NODE:"+left+":"+right))
		}
		layer = next
	}
	return layer[0], nil
}
