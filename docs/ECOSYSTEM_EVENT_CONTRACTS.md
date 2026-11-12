# Unified Payment Mesh — Treasury & Disbursement Event Contracts (`v1.2.0`)

`payout-engine` operates as the **Multi-Rail Treasury Clearing & 2PC Disbursement Gateway** within the four-repository financial core architecture:

```text
+-------------------+   FinancialEventEnvelope   +-------------------+
|   checkout-core   | -------------------------> |   ledger-vault    |
| (Edge Auth & FX)  |                            | (Serializable CoA)|
+-------------------+                            +-------------------+
          |                                                ^
          | Capture Cap & FX Pin                           | Settlement & Reversal
          v                                                v
+-------------------+   Waterfall Clawback Hook  +-------------------+
|   refund-guard    | -------------------------> |   payout-engine   |
| (Disputes & CE3)  |                            | (2PC Multi-Rail)  |
+-------------------+                            +-------------------+
```

---

## 1. Outbound Events (`payout-engine` $\rightarrow$ `ledger-vault`)

Emitted via `pkg/outbox/ledger_emitter.go` using the canonical `v1.2.0` pipe-delimited SHA-256 digest and `HMAC-SHA256` signature:

| Event Type | Trigger Stage | Debit Account (CoA) | Credit Account (CoA) |
| :--- | :--- | :--- | :--- |
| `payout.batch.committed` | Phase-2 2PC Lock & Merkle Seal | `LIABILITY:MERCHANT_PAYABLE_<CCY>` | `LIABILITY:DISBURSEMENT_IN_TRANSIT_<CCY>` |
| `payout.rail.settled` | `pacs.002` `ACCC` / ACH Settlement | `LIABILITY:DISBURSEMENT_IN_TRANSIT_<CCY>` | `ASSET:NOSTRO_CLEARING_<CCY>` |
| `payout.rail.returned` | NACHA `R01-R29` / `pacs.002` `RJCT` | `LIABILITY:DISBURSEMENT_IN_TRANSIT_<CCY>` | `LIABILITY:MERCHANT_PAYABLE_<CCY>` |
| `payout.reserve.clawed_back` | Waterfall Dispute Deduction | `LIABILITY:MERCHANT_RESERVE_<CCY>` | `ASSET:DISPUTE_RECOVERY_CLEARING_<CCY>` |

---

## 2. Inbound Directives (`refund-guard` $\rightarrow$ `payout-engine`)

When `refund-guard` processes an unrecovered partial refund or card-network `FIRST_CHARGEBACK`, it dispatches a signed `ClawbackDirective` into `pkg/clawback/offset_engine.go`, deducting obligations in strict conservation order:

1. **Matured Rolling Reserve** (`MaturedReserveMinor`)
2. **Pending Intraday Payable** (`PendingPayableMinor` — netted before the next banking cut-off)
3. **Locked Rolling Reserve** (`LockedReserveMinor` — when `AllowLockedReserveTap = true`)
4. **Residual Direct Debit Recovery** (`ResidualDebitRecoveryMinor`)
