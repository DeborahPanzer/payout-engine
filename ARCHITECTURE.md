# Payout Engine System Architecture

## 1. Two-Phase Commit (2PC) State Machine

All batch disbursements transition deterministically through `INITIATED -> PREPARED -> COMMITTED` (or `ABORTED`). No external bank wire (`FedNow`, `NACHA ACH`, `SEPA`) is dispatched unless all internal participants (`Ledger`, `OFAC`, `Liquidity`, `Tax`) vote `PREPARED`.

## 2. Conservation of Marketplace Funds

For every gross settlement $G$, platform fee $F$, tax withholding $T$, rolling reserve $R$, and connected recipient payouts $\sum P_i$:

$$G = F + T + R + \sum_{i=1}^{n} P_i$$

All arithmetic is executed strictly in 64-bit integer minor units (`int64`) to eliminate IEEE-754 floating-point drift.
