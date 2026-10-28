# Security & Treasury Protection Policy

## Supported Versions

| Version | Supported | Compliance Baseline |
| :--- | :--- | :--- |
| `1.1.x` | ✅ Yes | OFAC SDN / NACHA / ISO-20022 / PCI-DSS v4.0.1 |
| `1.0.x` | ✅ Yes | Security patches only |

## Treasury Isolation Controls

* **Zero-Trust Beneficiary Changes:** All bank routing/IBAN updates trigger a mandatory 24-hour cooling lock (`pkg/velocity`) to prevent Account Takeover (ATO) treasury draining.
* **Cryptographic Batch Integrity:** Outbound clearing files are sealed with SHA-256 Merkle roots (`pkg/merkle`) and signed via constant-time HMAC-SHA256 webhooks (`pkg/webhook`).
