# Changelog

All notable changes to `payout-engine` are documented in this file.
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-10-28

### Added
* **Two-Phase Commit Coordinator (`pkg/twophase`):** Unanimous prepare/commit batch disbursement state machine.
* **Rolling Reserve Engine (`pkg/reserve`):** Basis-point holdbacks and maturity window calculation.
* **Marketplace Split Engine (`pkg/split`):** Multi-party split settlement with exact penny-remainder conservation.
* **Multi-Rail Clearing (`pkg/ach`, `pkg/rtp`, `pkg/sepa`):** NACHA 94-byte CCD generator, FedNow ISO-20022 `pacs.008` builder, and IBAN Mod-97 validator.
* **Compliance & Risk (`pkg/ofac`, `pkg/tax`, `pkg/velocity`):** OFAC SDN screening, IRS Form 1099-K 24% backup withholding, and 24-hour bank-swap cooling lock.
* **Formal Verification (`specs/`):** TLA+ specifications for single-disbursement atomicity and reserve conservation.

## [1.0.0] - 2026-10-15

### Added
* Initial Go runtime batch disbursement coordinator and distributed retry queue specification.
