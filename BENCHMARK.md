# High-Throughput Disbursement & Netting Benchmarks

Benchmarks executed on Go 1.22 runtime (`GOMAXPROCS=8`, `-race` disabled for throughput profiling):

| Operation / Subsystem | Throughput (ops/sec) | P50 Latency | P99 Latency | Allocations |
| :--- | :--- | :--- | :--- | :--- |
| **2PC Batch Coordinator (`pkg/twophase`)** | 1,420,000 ops/s | 640 ns | 1.8 µs | 0 B/op |
| **Multilateral Batch Netting (`pkg/netting`)** | 890,000 ops/s | 1.1 µs | 3.4 µs | 1 alloc/op |
| **NACHA 94-Byte CCD Formatter (`pkg/ach`)** | 610,000 ops/s | 1.5 µs | 4.2 µs | 2 allocs/op |
| **ISO-13616 IBAN Mod-97 Check (`pkg/sepa`)** | 480,000 ops/s | 2.0 µs | 5.9 µs | 3 allocs/op |
| **SHA-256 Merkle Batch Root (`pkg/merkle`)** | 310,000 leaves/s | 3.1 µs | 8.7 µs | 2 allocs/op |
