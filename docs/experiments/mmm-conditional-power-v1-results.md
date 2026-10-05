# Conditional Anti-Pigeon power v1: late-shift failure

The [frozen feasibility protocol](mmm-conditional-power-v1-protocol.md) fails
in both design and independent confirmation. Its cumulative two-context
certificate reliably flags a swap present from the first live frame when a
512-label reference batch exists, but almost never flags the same swap when
it begins at clock 256. This is a direct transfer limit of the earlier toy
result, not a failure of the separate MMM implementation.

## Confirmation

Each cell contains 1,000 independent 512-clock trajectories. Reference and
live streams are independently sampled, and the 25% audit, missingness and
delay draws are independent of outcomes. Counts include only arrived audits.
The batch-reference strategy spends 512 pre-live labels; online reference
uses about 128 immediate or 100 sparse-delayed arrived labels by horizon.

| Reference | Schedule | Stable flags | Start-swap flags; median clock | Clock-256 swap flags |
| --- | --- | ---: | ---: | ---: |
| Batch 512 | Immediate | 0 | 1,000; 134 | **16** |
| Batch 512 | Sparse delayed | 0 | 1,000; 182 | **1** |
| Online | Immediate | 0 | 994; 352 | **0** |
| Online | Sparse delayed | 0 | 778; 435 | **0** |

All clock-256 flags occur after the change. The online sparse-delayed
start-swap arm also misses the predeclared 800/1,000 power floor. In the
batch-reference immediate arm, the 16 clock-256 detections have median
clock 497, leaving almost no horizon for useful downstream adaptation.
The design split agrees: 11/1,000 and 1/1,000 clock-256 detections with a
batch reference; zero in both online-reference schedules. These figures do
not estimate performance for weaker shifts, dependent sources, learned
contexts or ongoing graph changes.

The explanation is testable rather than assumed: the live prefix before clock
256 has the original conditional law. A cumulative mean near horizon is a
mixture of old and new outcomes, reducing an 0.8 per-context law gap to about
0.4 in expectation, while the frozen simultaneous radius still charges for
uncertainty. The online-reference arm also spends uncertainty on its smaller
reference sample. A windowed or restarted audit is needed to test the
prefix-dilution hypothesis, with fresh seeds and its own repeated-testing
accounting; simply lowering the threshold would sacrifice the stated false-
revocation control.

## Audit and scope

- [Design rows](mmm-conditional-power-v1-design.jsonl), SHA256
  `3db31bce43b96c26308498e781478cfed1533cacbef7f13639ed5f35a0643502`.
- [Confirmation rows](mmm-conditional-power-v1-confirmation.jsonl), SHA256
  `ba65a5d51c033de5a7bf3c7c386ad98d88f2a8ee16e43a1f85325f3900c3d9c4`.
- [Simulator](../../research/conditional-power-v1.mjs) and
  [independent verifier](../../research/conditional-power-v1-verify.mjs)
  preserve all 24,000 trial rows. The verifier passes source/protocol hashes,
  every unique row, evidence accounting and all 24 summary cells. A fresh
  confirmation replay matches all 12,002 header/trial/summary rows exactly
  after excluding elapsed wall time.
- The full 12,000-trajectory confirmation simulation took 1.54 s, including
  random generation, queueing, gate checks, assertions and JSON output.
  This is not a Go daemon or loaded serving benchmark.

The false-flag bound is conditional on iid bounded labels within each fixed
context and independent audit/arrival selection. It says nothing about
unknown external-law diameter, multiple active buckets, or a learned
partition. The actual MMM split path also requires a scalar Bayesian
nomination; this component does not bypass or validate that requirement.
No forecast law is changed here, so Goal 3's downstream Brier requirement
remains untested. All seven whole goals stay open; production is untouched.
