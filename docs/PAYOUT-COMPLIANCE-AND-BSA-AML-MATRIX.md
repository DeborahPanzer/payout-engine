# Regulatory & Banking Compliance Control Matrix

This matrix maps the controls inside `payout-engine` to **US FinCEN BSA/AML**, **OFAC**, **NACHA Operating Rules**, **ISO-20022**, and **IRS Internal Revenue Code** mandates.

| Regulatory Standard | Control Requirement | Implementing Package | Technical Enforcement |
| :--- | :--- | :--- | :--- |
| **OFAC SDN Sanctions** | Block disbursements to Specially Designated Nationals | `pkg/ofac` | Pre-disbursement entity normalization and exact/substring blocklist interception. |
| **NACHA Operating Rules** | Validate Receiver DFI routing numbers & block R02/R03 retries | `pkg/ach` & `pkg/returncodes` | Weighted Modulo-10 ABA checksum validation and terminal return code account invalidation. |
| **ISO-20022 (`pacs.008`)** | Structured interbank real-time credit transfer messaging | `pkg/rtp` | Deterministic XML builder for FedNow and TCH Real-Time Payments (RTP). |
| **ISO-13616 (SEPA/SWIFT)** | International Bank Account Number check digit verification | `pkg/sepa` | Arbitrary-precision `big.Int` Modulo-97 validation prior to cross-border wire dispatch. |
| **IRS IRC § 3406 / 1099-K** | Statutory 24% backup withholding on unverified TIN sellers | `pkg/tax` | Automated YTD gross volume tracking and 2,400 bps tax deduction before net payout. |
