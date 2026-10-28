package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func SignPayoutWebhookPayload(timestampUnix int64, payloadJSON string, secretKey string) string {
	canonical := fmt.Sprintf("%d.%s", timestampUnix, payloadJSON)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(canonical))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", timestampUnix, sig)
}

func VerifyPayoutWebhookSignature(timestampUnix int64, payloadJSON, headerSig, secretKey string) bool {
	expected := SignPayoutWebhookPayload(timestampUnix, payloadJSON, secretKey)
	return hmac.Equal([]byte(expected), []byte(headerSig))
}
