# ADR 0002: Multi-Rail Bank Clearing (FedNow, NACHA ACH, SEPA) and OFAC Pre-Screening

* **Status:** Accepted
* **Date:** 2026-10-28
* **Deciders:** Deborah Panzer (Principal Payment Systems Architect)

## Context and Problem Statement

Global marketplaces require both instant 24/7 liquidity (`ISO-20022 pacs.008` via FedNow/RTP) and low-cost bulk clearing (94-byte NACHA ACH and SEPA Credit Transfers), while strictly enforcing OFAC SDN sanctions and IRS 1099-K backup withholding.

## Decision Outcome

We decouple payout orchestration from underlying clearing protocols using rail-specific generators (`pkg/ach`, `pkg/rtp`, `pkg/sepa`) guarded by a pre-execution compliance pipeline (`pkg/ofac`, `pkg/tax`) and automatic circuit-breaker rail failover (`pkg/circuit`).
