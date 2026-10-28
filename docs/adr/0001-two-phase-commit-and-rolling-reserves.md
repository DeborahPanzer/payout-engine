# ADR 0001: Two-Phase Commit (2PC) Batch Coordination and Rolling Reserve Holdbacks

* **Status:** Accepted
* **Date:** 2026-10-28
* **Deciders:** Deborah Panzer (Principal Payment Systems Architect)

## Context and Problem Statement

Outbound marketplace disbursements are irreversible once settled over external banking rails (FedNow, RTP, NACHA ACH, SEPA). A partial failure between internal ledger debiting, OFAC sanctions screening, and bank file dispatch can cause catastrophic double-payouts or uncollateralized chargeback exposure.

## Decision Outcome

1. **Unanimous Two-Phase Commit (`pkg/twophase`):** Every disbursement batch executes a strict `PREPARE` phase across Ledger, OFAC Compliance, and Liquidity Pool nodes before transitioning to `COMMITTED`.
2. **Automated Rolling Reserves (`pkg/reserve`):** Configurable basis-point holdbacks are isolated in merchant reserve escrow for $N$ days to collateralize downstream dispute reversals from `refund-guard`.
