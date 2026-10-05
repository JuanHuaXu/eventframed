# Conditional gate transport v1: frozen diagnostic protocol

This study tests direct transport of the finite-horizon two-cell Hoeffding
gate from conditional-gate v2 to the observation volume and outcome rules of
the MMM member experiment. It is an isolated diagnostic, not actual MMM
nomination, a target-law diameter certificate, or a proposal to replace the
existing Anti-Pigeon gate. Do not tune the gate on these streams.

Use 1,000 independent trajectories per case/schedule/split, each 512 clocks.
Design and confirmation seed bases are `2026100131` and `2026100132`.
Each clock has independent uniform nine-bit live and reference contexts and
an outcome-blind shared 25% full-frame audit. Before any shift, both labels
use parity of bits 6, 7, and 8 with 5% independent flip noise. A shifted side
uses bit 2 with the same noise. Cases: stable; live-only shift at clock 256;
common shift at clock 256; and live-only shift from clock 0. The declared
gate context is bit 2 of each side's independently sampled frame.

Schedules: immediate complete audited labels, as in the original member
experiment; and stress with independent 20% missingness per audited side and
integer-uniform 0..31 delivery delay. Each label enters the gate only when
its own packet arrives. Neither side sees missing, future, or unaudited
outcomes. Track per-cell counts, flags, pending deliveries, and total audit
requests. A separate random role drives each source and schedule decision.

Use the unchanged v2 rule with `T=512`, `delta=0.02`, `epsilon=0.10` and
`r(n)=sqrt(log(4*2*(T+1)/delta)/(2*n))`. At each clock after delivery,
flag when any cell with observations on both sides has

`abs(mean_live - mean_ref) > epsilon + r(n_live) + r(n_ref)`.

The direct transport passes this *component power screen* only if, on both
splits and schedules, it flags at most 20/1,000 stable and common-shift
trajectories and at least 800/1,000 of each live-shift case by clock 511.
Report median flag clock, mean per-cell label counts, audit count, and
pending count even on failure. Common nonstationary drift with independently
arriving labels is a stress control, not covered by the stationary
two-sample Hoeffding argument. This screen does not establish simultaneous
multi-bucket error control even if it passes.

Preserve every row, protocol and source hashes, an independent row-based
summary, and deterministic replay. A failed screen rejects only direct
transport of this gate at this evidence volume. It does not reject
variance-adaptive betting, different context partitions, matched-pair
evidence, more audits, or a valid precollected reference with its cost.
