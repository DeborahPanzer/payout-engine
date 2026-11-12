# ADR-0003: Monotonic Fencing Tokens & Asynchronous Bank Return Compensation

* **Status:** Accepted (`v1.2.0`)
* **Components:** `internal/coordinator`, `pkg/pacs002`, `pkg/noc`, `pkg/returncodes`, `pkg/outbox`

---

## 1. Context & Problem Statement

Distributed disbursement gateways face two catastrophic failure modes in production:

1. **Split-Brain Double Disbursement:** A worker holding a 2PC batch lock experiences a stop-the-world GC pause or network partition past its lease TTL. A failover worker acquires the batch and submits the FedNow/ACH wire; moments later, the stalled worker resumes and dispatches a duplicate wire.
2. **Asynchronous Post-Commit Bank Rejections:** Even after a batch passes 2PC and is transmitted to the clearing house, the receiving DFI may asynchronously reject the credit hours or days later via **NACHA Returns (`R01–R29`)**, **Notification of Change (`C01–C07`)**, or **ISO-20022 `pacs.002` (`RJCT`)**.

## 2. Architectural Decisions

1. **Monotonic Epoch Fencing (`internal/coordinator`):**
   * Every lease acquisition increments a strictly monotonic `uint64` `FencingEpoch`.
   * Bank rail adapters and the transactional outbox MUST call `ValidateFencingToken(batchID, epoch, now)` immediately prior to wire transmission, deterministically rejecting stale zombie epochs (`ErrStaleFencingToken`).
2. **Append-Only Compensating Reversals (`pkg/pacs002` & `pkg/outbox`):**
   * Historical committed batches are **never** mutated or deleted in place.
   * Terminal rejections (`RJCT` / `R01–R29`) emit an append-only `payout.rail.returned` compensating journal entry to `ledger-vault`.
   * Regulatory rejections (`RR04`) and invalid-account rejections (`AC04`, `R02`, `R03`) atomically freeze the beneficiary bank profile while suppressing automated failover retries.
3. **Automated 6-Banking-Day NOC Remediation (`pkg/noc`):**
   * Nacha Addenda Type `98` (`C01–C07`) updates are validated against Modulo-10 ABA checksums before updating beneficiary routing profiles within the mandatory 6-banking-day Nacha compliance window.
