# Operational Runbook: FedNow / RTP Switch Degradation & Same-Day ACH Failover

## 1. Trigger Conditions

Alert `PAYOUT_INSTANT_RAIL_CIRCUIT_OPEN` fires when:

* Consecutive FedNow / RTP ISO-20022 `pacs.008` submissions time out or return `pacs.002` `RJCT` with transient switch codes (`AB05`, `TM01`) exceeding the `pkg/circuit` failure threshold.
* Primary Nostro instant-settlement pool buffer drops below the minimum reserve floor in `pkg/liquidity`.

## 2. Automated System Behavior

1. `pkg/circuit` trips the instant rail to `OPEN` and blocks unacknowledged retries until `ValidateFencingToken` confirms lease exclusivity.
2. `pkg/pacs002` classifies `AB05` / `TM01` as `REVERSE_HOLD_ELIGIBLE_FOR_FAILOVER` (`TriggerRailFailover = true`).
3. `pkg/calendar` evaluates Eastern Time against Federal Reserve Same-Day ACH Windows (`10:30`, `14:45`, `16:45 ET`) and routes eligible USD batches to `pkg/ach` without changing the root `IdempotencyKey`.

## 3. Standard Operating Procedure (SOP)

1. **Step 1 — Verify `pacs.002` In-Flight Status (`ACSP` / `PDNG`):**
   * **NEVER** fail over a disbursement whose latest status is `IN_FLIGHT_PENDING_HOLD` (`ACSP` or `PDNG`). Wait for terminal FedNow status confirmation or execute a formal FedNow status inquiry (`pacs.028`) first.
2. **Step 2 — Confirm Fencing Epoch Exclusivity:**
   * Inspect `internal/coordinator` metrics to confirm zero `ErrStaleFencingToken` alerts before draining queued failover batches.
3. **Step 3 — Reconcile Nostro Pool Reservations:**
   * Verify that released instant-pool reservations match newly allocated ACH Nostro batch totals and `ledger-vault` `DISBURSEMENT_IN_TRANSIT` balances.
