package webhook

import "testing"

func TestSignAndVerifyPayoutWebhook(t *testing.T) {
	payload := `{"event":"payout.paid","payout_id":"po_881"}`
	header := SignPayoutWebhookPayload(1761619200, payload, "whsec_test_key")
	if !VerifyPayoutWebhookSignature(1761619200, payload, header, "whsec_test_key") {
		t.Fatal("expected valid HMAC-SHA256 webhook signature to verify")
	}
}
