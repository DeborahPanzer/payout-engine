<div align="center">

# 💸 Payout Engine & Treasury Gateway

[![Payout Engine CI & Compliance Suite](https://github.com/DeborahPanzer/payout-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/DeborahPanzer/payout-engine/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Architecture: Two-Phase Commit | Queued](https://img.shields.io/badge/Architecture-Two--Phase_Commit_%7C_Queued-2ea44f)](#)
[![Compliance: OFAC | NACHA | ISO-20022](https://img.shields.io/badge/Compliance-OFAC_%7C_NACHA_%7C_ISO--20022-blueviolet)](#)
[![Formal Verification: TLA+](https://img.shields.io/badge/Formal_Verification-TLA%2B-informational)](#)
[![Language: Go 100%](https://img.shields.io/badge/Language-Go_100%25-00ADD8)](#)
[![Sponsor](https://img.shields.io/badge/Sponsor-DeborahPanzer-ea4aaa?logo=github-sponsors&logoColor=white)](https://github.com/sponsors/DeborahPanzer)

</div>

---

`payout-engine` is an enterprise-grade, zero-dependency automated disbursement and treasury settlement gateway built for **Go (Golang)** runtimes. It orchestrates **Two-Phase Commit (2PC) batch payouts**, **multi-party marketplace splits**, **rolling reserve holdbacks**, **OFAC sanctions screening**, and **multi-rail bank clearing (FedNow ISO-20022, NACHA ACH, SEPA)**.

---

## 🏛️ Architectural Disbursement Pipeline

```text
                  [Settlement Trigger / Split Request]
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────┐
│ 1. COMPLIANCE, TAX & ATO VELOCITY SHIELD                         │
│    • OFAC SDN Sanctions Screening Gate (pkg/ofac)                │
│    • 24h Beneficiary Bank-Swap Cooling Guard (pkg/velocity)      │
│    • IRS Form 1099-K & 24% Backup Withholding (pkg/tax)          │
└──────────────────────────────────┬───────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────┐
│ 2. SPLIT ACCOUNTING, NETTING & 2PC COORDINATION                  │
│    • Multi-Party Marketplace Split Engine (pkg/split)            │
│    • Rolling Reserve Holdback & Maturity Lock (pkg/reserve)      │
│    • Multilateral Intraday Obligation Netting (pkg/netting)      │
│    • Unanimous Two-Phase Commit Coordinator (pkg/twophase)       │
└──────────────────────────────────┬───────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────┐
│ 3. TREASURY LIQUIDITY & MULTI-RAIL BANK CLEARING                 │
│    • Nostro Pool Liquidity & Buffer Router (pkg/liquidity)       │
│    • FedNow / RTP ISO-20022 pacs.008 XML Builder (pkg/rtp)       │
│    • US NACHA ACH 94-Byte CCD & ABA Mod-10 Engine (pkg/ach)      │
│    • SEPA ISO-13616 IBAN Modulo-97 Validator (pkg/sepa)          │
│    • SHA-256 Merkle Batch Sealer & DLQ Retry Queue (pkg/merkle)  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 📂 Repository Structure & Core Packages

```text
payout-engine/
├── .github/workflows/ci.yml                   # Automated Go vet & concurrency race detector CI
├── docs/
│   ├── adr/
│   │   ├── 0001-two-phase-commit-and-rolling-reserves.md # ADR-0001: 2PC & rolling reserves
│   │   └── 0002-multi-rail-clearing-fednow-ach-sepa.md   # ADR-0002: Multi-rail clearing
│   └── PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md # OFAC, NACHA, ISO-20022 & IRS control matrix
├── pkg/
│   ├── ach/                                   # NACHA 94-byte CCD generator & ABA Mod-10 check
│   ├── circuit/                               # Bank rail circuit breaker (FedNow -> ACH failover)
│   ├── liquidity/                             # Nostro treasury pool selector & buffer guard
│   ├── merkle/                                # SHA-256 Merkle root batch manifest sealer
│   ├── netting/                               # Multilateral batch obligation netting engine
│   ├── ofac/                                  # OFAC SDN sanctions entity screening gate
│   ├── queue/                                 # Exponential backoff retry scheduler & DLQ isolation
│   ├── reserve/                               # Merchant rolling reserve & maturity calculator
│   ├── returncodes/                           # NACHA ACH return classifier (R01-R29 dispositions)
│   ├── rtp/                                   # FedNow / RTP ISO-20022 (pacs.008.001.08) builder
│   ├── sepa/                                  # ISO-13616 arbitrary-precision IBAN Mod-97 check
│   ├── split/                                 # Marketplace multi-party split & penny allocation
│   ├── tax/                                   # US IRS Form 1099-K & 24% backup withholding
│   ├── twophase/                              # Atomic Two-Phase Commit (2PC) batch coordinator
│   ├── velocity/                              # Bank-swap cooling lock & daily velocity limiter
│   └── webhook/                               # Constant-time HMAC-SHA256 webhook dispatcher
├── specs/
│   ├── RollingReserveInvariant.tla            # TLA+ formal proof of reserve conservation
│   └── TwoPhasePayoutCommit.tla               # TLA+ formal proof of single-wire atomicity
├── .env.example                               # Treasury & 2PC environment variable template
├── .gitattributes                             # Linguist 100% Go classification & LF rules
├── .gitignore                                 # Go binary & secret keystore exclusions
├── ARCHITECTURE.md                            # Formal 2PC state machines & split conservation
├── BENCHMARK.md                               # Go runtime throughput & P99 latency telemetry
├── CHANGELOG.md                               # Enterprise semantic release history (v1.1.0)
├── SECURITY.md                                # Treasury isolation & ATO disclosure policy
├── engine.go                                  # Root batch disbursement coordinator entrypoint
└── go.mod                                     # Go module definition (zero external dependencies)
```

---

## ⚙️ Key Enterprise Capabilities

* **Atomic Two-Phase Commit (2PC) & Rolling Reserves:** Guarantees unanimous readiness across ledger, compliance, and treasury nodes (`pkg/twophase`) while isolating configurable basis-point chargeback reserves (`pkg/reserve`).
* **Multi-Rail Clearing (FedNow, NACHA ACH, SEPA):** Generates **ISO-20022 (`pacs.008.001.08`)** instant credit transfers (`pkg/rtp`), **94-byte NACHA CCD** batch records with Modulo-10 ABA routing verification (`pkg/ach`), and **ISO-13616 Modulo-97** validated IBAN instructions (`pkg/sepa`).
* **Pre-Disbursement OFAC & IRS 1099-K Governance:** Screens beneficiaries against **OFAC SDN** sanctions blocklists (`pkg/ofac`), enforces 24-hour cooling locks after bank routing updates (`pkg/velocity`), and applies statutory **24% IRS backup withholding** on unverified TIN accounts (`pkg/tax`).
* **Multilateral Netting & Nostro Liquidity Routing:** Compresses intraday payables and clawbacks into net settlement positions (`pkg/netting`) and selects cost-optimal **Nostro treasury pools** without breaching minimum liquidity buffers (`pkg/liquidity`).

---

## 🧪 Verification & Concurrency Race Suite

```bash
# 1. Run static analysis across all disbursement packages
go vet ./...

# 2. Execute unit test suite with Go concurrency race detector enabled
go test -v -race ./...
```

---

## 📜 Documentation & Governance

* [**System Architecture & 2PC State Machines (`ARCHITECTURE.md`)**](ARCHITECTURE.md)
* [**OFAC, NACHA & BSA/AML Compliance Matrix (`docs/PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md`)**](docs/PAYOUT-COMPLIANCE-AND-BSA-AML-MATRIX.md)
* [**ADR-0001: 2PC Coordination & Rolling Reserves (`docs/adr/0001-two-phase-commit-and-rolling-reserves.md`)**](docs/adr/0001-two-phase-commit-and-rolling-reserves.md)
* [**ADR-0002: Multi-Rail Clearing & OFAC Screening (`docs/adr/0002-multi-rail-clearing-fednow-ach-sepa.md`)**](docs/adr/0002-multi-rail-clearing-fednow-ach-sepa.md)
* [**Throughput & Concurrency Benchmarks (`BENCHMARK.md`)**](BENCHMARK.md)
* [**Treasury Security Policy (`SECURITY.md`)**](SECURITY.md)

## License

Licensed under the **Apache License, Version 2.0**. See [LICENSE](LICENSE) for full details.
