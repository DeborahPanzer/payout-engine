<div align="center">

# 💸 Payout Engine & Treasury Gateway

[![Payout Engine CI & Compliance Suite](https://github.com/DeborahPanzer/payout-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/DeborahPanzer/payout-engine/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Architecture: Two-Phase Commit | Queued](https://img.shields.io/badge/Architecture-Two--Phase_Commit_%7C_Queued-2ea44f)](ARCHITECTURE.md)
[![Compliance: OFAC | NACHA | ISO-20022](https://img.shields.io/badge/Compliance-OFAC_%7C_NACHA_%7C_ISO--20022-blueviolet)](docs/PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md)
[![Formal Verification: TLA+](https://img.shields.io/badge/Formal_Verification-TLA%2B-informational)](specs/TwoPhasePayoutCommit.tla)
[![Language: Go 100%](https://img.shields.io/badge/Language-Go_100%25-00ADD8)](go.mod)
[![Sponsor](https://img.shields.io/badge/Sponsor-DeborahPanzer-ea4aaa?logo=github-sponsors&logoColor=white)](https://github.com/sponsors/DeborahPanzer)

</div>

---

`payout-engine` is an enterprise-grade, zero-dependency automated disbursement and treasury settlement gateway built for **Go (Golang)** runtimes. It orchestrates **fenced Two-Phase Commit (2PC) batch payouts**, **multi-party marketplace splits**, **rolling reserve & waterfall clawbacks**, **OFAC sanctions screening**, and **bidirectional multi-rail bank clearing (FedNow ISO-20022, NACHA ACH, SEPA)**.

---

## 🏛️ Architectural Disbursement Pipeline

```text
                        [Settlement Trigger / Split Request]
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ 1. COMPLIANCE, TAX, BANKING CUT-OFF & ATO VELOCITY SHIELD                │
│    • OFAC SDN Sanctions Screening Gate (pkg/ofac)                        │
│    • 24h Beneficiary Bank-Swap Cooling Guard (pkg/velocity)              │
│    • IRS Form 1099-K & 24% Backup Withholding (pkg/tax)                  │
│    • Fedwire / Same-Day ACH / TARGET2 Cut-Off Router (pkg/calendar)      │
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ 2. SPLIT ACCOUNTING, CLAWBACK WATERFALL & FENCED 2PC COORDINATION        │
│    • Multi-Party Marketplace Split Engine (pkg/split)                    │
│    • Rolling Reserve Holdback & Dispute Clawback (pkg/reserve, clawback) │
│    • Multilateral Intraday Obligation Netting (pkg/netting)              │
│    • Fenced Two-Phase Commit Coordinator (pkg/twophase, internal/)       │
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────┐
│ 3. MULTI-RAIL CLEARING, RETURN COMPENSATION & ECOSYSTEM OUTBOX           │
│    • Nostro Pool Liquidity & Rail Circuit Breaker (pkg/liquidity)        │
│    • FedNow ISO-20022 pacs.008 & pacs.002 Status (pkg/rtp, pkg/pacs002)  │
│    • NACHA ACH CCD, R-Returns & NOC C01-C07 (pkg/ach, returncodes, noc)  │
│    • SEPA ISO-13616 IBAN Modulo-97 Validator (pkg/sepa)                  │
│    • HMAC Ledger Outbox & SHA-256 Merkle Sealer (pkg/outbox, pkg/merkle) │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 🔗 Unified Payment Mesh Integration (`v1.2.0`)

`payout-engine` acts as the outbound treasury settlement and post-settlement return engine across the four-service financial core ecosystem, communicating via replay-protected `HMAC-SHA256` `FinancialEventEnvelope` contracts (`pkg/outbox/ledger_emitter.go`):

* **`ledger-vault`:** Ingests `payout.batch.committed`, `payout.rail.settled`, `payout.rail.returned`, and `payout.reserve.clawed_back` postings into an append-only, `Serializable` ACID double-entry Chart of Accounts.
* **`refund-guard`:** Dispatches signed dispute and over-capture `ClawbackDirective` instructions into `pkg/clawback`, offsetting matured rolling reserves and intraday payables before bank wire dispatch.
* **`checkout-core`:** Supplies upstream merchant settlement credits, FX rate pins, and Interchange++ fee splits that feed intraday multilateral netting (`pkg/netting`).

---

## 📂 Repository Structure & Core Packages

```text
payout-engine/
├── .github/
│   ├── FUNDING.yml                                    # GitHub Sponsors manifest
│   └── workflows/ci.yml                               # Automated Go vet & concurrency race detector CI
├── docs/
│   ├── adr/
│   │   ├── 0001-two-phase-commit-and-rolling-reserves.md # ADR-0001: 2PC & rolling reserves
│   │   ├── 0002-multi-rail-clearing-fednow-ach-sepa.md   # ADR-0002: Multi-rail clearing
│   │   └── 0003-fencing-tokens-and-pacs002-return-compensation.md # ADR-0003: Fencing & returns
│   ├── runbooks/
│   │   ├── nacha-return-noc-and-nostro-reconciliation.md # SOP: NACHA R/C-codes & RR04 freeze
│   │   └── rail-outage-fednow-to-ach-failover.md      # SOP: Instant rail outage & ACH failover
│   ├── ECOSYSTEM_EVENT_CONTRACTS.md                   # v1.2.0 cross-service HMAC event contract
│   └── PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md        # OFAC, NACHA, ISO-20022 & IRS control matrix
├── internal/
│   └── coordinator/                                   # Monotonic fencing-token 2PC lease lock guard
├── pkg/
│   ├── ach/                                           # NACHA 94-byte CCD generator & ABA Mod-10 check
│   ├── calendar/                                      # FedACH Same-Day windows, holidays & TARGET2 router
│   ├── circuit/                                       # Bank rail circuit breaker (FedNow -> ACH failover)
│   ├── clawback/                                      # Waterfall reserve & payable dispute offset engine
│   ├── liquidity/                                     # Nostro treasury pool selector & buffer guard
│   ├── merkle/                                        # SHA-256 Merkle root batch manifest sealer
│   ├── netting/                                       # Multilateral batch obligation netting engine
│   ├── noc/                                           # NACHA Addenda 98 NOC (C01-C07) parser & 6-day SLA
│   ├── ofac/                                          # OFAC SDN sanctions entity screening gate
│   ├── outbox/                                        # HMAC-SHA256 ecosystem ledger event outbox
│   ├── pacs002/                                       # ISO-20022 (pacs.002.001.10) status report evaluator
│   ├── queue/                                         # Exponential backoff retry scheduler & DLQ isolation
│   ├── reserve/                                       # Merchant rolling reserve & maturity calculator
│   ├── returncodes/                                   # NACHA ACH return classifier (R01-R29 dispositions)
│   ├── rtp/                                           # FedNow / RTP ISO-20022 (pacs.008.001.08) builder
│   ├── sepa/                                          # ISO-13616 arbitrary-precision IBAN Mod-97 check
│   ├── split/                                         # Marketplace multi-party split & penny allocation
│   ├── tax/                                           # US IRS Form 1099-K & 24% backup withholding
│   ├── twophase/                                      # Atomic Two-Phase Commit (2PC) batch coordinator
│   ├── velocity/                                      # Bank-swap cooling lock & daily velocity limiter
│   └── webhook/                                       # Constant-time HMAC-SHA256 webhook dispatcher
├── specs/
│   ├── RollingReserveInvariant.tla                    # TLA+ formal proof of reserve conservation
│   └── TwoPhasePayoutCommit.tla                       # TLA+ formal proof of single-wire atomicity
├── .env.example                                       # Treasury & 2PC environment variable template
├── .gitattributes                                     # Linguist 100% Go classification & LF rules
├── .gitignore                                         # Go binary & secret keystore exclusions
├── .golangci.yml                                      # Strict Go financial error & security linter rules
├── ARCHITECTURE.md                                    # Formal 2PC state machines & split conservation
├── BENCHMARK.md                                       # Go runtime throughput & P99 latency telemetry
├── CHANGELOG.md                                       # Enterprise semantic release history
├── LICENSE                                            # Apache-2.0 open-source license
├── README.md                                          # Architectural overview & operational index
├── SECURITY.md                                        # Treasury isolation & ATO disclosure policy
├── engine.go                                          # Root batch disbursement coordinator entrypoint
└── go.mod                                             # Go module definition (zero external dependencies)
```

---

## ⚙️ Key Enterprise Capabilities

* **Fenced Two-Phase Commit (2PC), Rolling Reserves & Waterfall Clawbacks:** Guarantees unanimous readiness across ledger, compliance, and treasury nodes (`pkg/twophase`, `internal/coordinator` monotonic fencing epochs) while isolating chargeback reserves (`pkg/reserve`) and executing conservation-safe dispute offsets (`pkg/clawback`).
* **Bidirectional Multi-Rail Clearing (FedNow, NACHA ACH, SEPA):** Generates **ISO-20022 (`pacs.008.001.08`)** instant transfers (`pkg/rtp`), evaluates asynchronous **`pacs.002.001.10`** settlement/rejection callbacks (`pkg/pacs002`), builds **94-byte NACHA CCD** batches with **`R01–R29`** return & **`C01–C07` Addenda 98 NOC** automation (`pkg/ach`, `pkg/returncodes`, `pkg/noc`), and validates **ISO-13616 Modulo-97** IBANs (`pkg/sepa`).
* **Banking Cut-Off Routing, OFAC & IRS 1099-K Governance:** Evaluates **Federal Reserve Same-Day ACH windows & holidays** (`pkg/calendar`) for automatic 24/7 FedNow promotion, screens **OFAC SDN** sanctions (`pkg/ofac`), enforces 24-hour bank-swap cooling locks (`pkg/velocity`), and applies statutory **24% IRS backup withholding** (`pkg/tax`).
* **Multilateral Netting, Nostro Liquidity & HMAC Ledger Sync:** Compresses intraday payables into net settlement positions (`pkg/netting`), protects **Nostro pool buffers** (`pkg/liquidity`), seals batches into **SHA-256 Merkle Trees** (`pkg/merkle`), and emits constant-time **HMAC-SHA256** signed journal envelopes (`pkg/outbox`).

---

## 🧪 Verification & Concurrency Race Suite

```bash
# 1. Copy treasury environment variable template for local verification
cp .env.example .env

# 2. Run strict static analysis across all disbursement & internal packages
go vet ./...

# 3. Execute full unit & concurrency test suite with Go race detector enabled
go test -v -race ./...
```

---

## 📜 Documentation, Runbooks & Governance

* [**System Architecture & 2PC State Machines (`ARCHITECTURE.md`)**](ARCHITECTURE.md)
* [**Unified Ecosystem Event Contracts (`docs/ECOSYSTEM_EVENT_CONTRACTS.md`)**](docs/ECOSYSTEM_EVENT_CONTRACTS.md)
* [**OFAC, NACHA & BSA/AML Compliance Matrix (`docs/PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md`)**](docs/PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md)
* [**ADR-0001: 2PC Coordination & Rolling Reserves (`docs/adr/0001-two-phase-commit-and-rolling-reserves.md`)**](docs/adr/0001-two-phase-commit-and-rolling-reserves.md)
* [**ADR-0002: Multi-Rail Clearing & OFAC Screening (`docs/adr/0002-multi-rail-clearing-fednow-ach-sepa.md`)**](docs/adr/0002-multi-rail-clearing-fednow-ach-sepa.md)
* [**ADR-0003: Fencing Tokens & `pacs.002` Compensation (`docs/adr/0003-fencing-tokens-and-pacs002-return-compensation.md`)**](docs/adr/0003-fencing-tokens-and-pacs002-return-compensation.md)
* [**Runbook: FedNow Outage & Same-Day ACH Failover (`docs/runbooks/rail-outage-fednow-to-ach-failover.md`)**](docs/runbooks/rail-outage-fednow-to-ach-failover.md)
* [**Runbook: NACHA Returns, NOC SLA & `RR04` Holds (`docs/runbooks/nacha-return-noc-and-nostro-reconciliation.md`)**](docs/runbooks/nacha-return-noc-and-nostro-reconciliation.md)
* [**Throughput & Concurrency Benchmarks (`BENCHMARK.md`)**](BENCHMARK.md)
* [**Enterprise Release Changelog (`CHANGELOG.md`)**](CHANGELOG.md)
* [**Treasury Security Policy (`SECURITY.md`)**](SECURITY.md)

## License

Licensed under the **Apache License, Version 2.0**. See [LICENSE](LICENSE) for full details.
