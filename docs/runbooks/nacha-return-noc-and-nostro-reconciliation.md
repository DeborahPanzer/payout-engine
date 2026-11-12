# Operational Runbook: NACHA Return / NOC Triage & Regulatory `RR04` Escalation

## 1. Trigger Conditions

This runbook governs three treasury exception queues:

* `NACHA_ADMIN_OR_UNAUTH_RETURN`: Receipt of Nacha return codes `R02` (Account Closed), `R03` (No Account), `R10` / `R29` (Unauthorized Corporate/Consumer Entry) in `pkg/returncodes`.
* `NACHA_NOC_SLA_WARNING`: Unapplied Addenda Type `98` (`C01–C07`) instructions in `pkg/noc` approaching the **6-banking-day** Nacha compliance deadline.
* `ISO20022_REGULATORY_REJECT_RR04`: Receipt of `pacs.002` `RJCT` with reason `RR04` (`REVERSE_HOLD_AND_ESCALATE_COMPLIANCE`).

## 2. Standard Operating Procedure (SOP)

1. **Step 1 — Regulatory Freeze (`RR04` / OFAC Match):**
   * Confirm `FreezeBeneficiaryAccount = true` has locked all outbound disbursements for the entity.
   * Escalate original beneficiary KYC/TIN and `pkg/ofac` screening logs to the BSA/AML Compliance Officer within 4 hours. Do **not** disclose sanctions rejection details to external webhooks.
2. **Step 2 — Remediate `ErrInvalidABARoutingChecksum` NOC Exceptions:**
   * If `pkg/noc` quarantines an incoming `C02`/`C03`/`C07` record due to Modulo-10 failure, request a signed ODFI clarification notice and place the merchant's next payout on temporary bank-verification hold.
3. **Step 3 — Reconcile Compensating Ledger Entries:**
   * Verify `pkg/outbox` has dispatched `payout.rail.returned` to `ledger-vault` and that `Acknowledge(eventID)` cleared the pending outbox queue before closing the incident.
