# Snapshot-specialist arithmetic and detached-law component

Status: component checks PASS; integrated policy and quality UNTESTED.
No promotion, serving change or inherited statistical guarantee.

Subsequent [gate/controller integration](snapshot-specialists-integration-results.md)
passes its scoped lifecycle tests and records the full journal cost. Quality
remains untested; the original component-only scope below is preserved.

## Frozen finite contract

Keep two publication banks, four models each, with immutable identity
1+4v+j for version v and role j. Publications occur at origins32v for v=0..7.
The bounded experiment has256 issues, at most64 uncommitted issue messages and
80 total message slots including publication transitions. The owner is externally
serialized; race-test success is not a claim of concurrent-method safety.

Initial role prior is (.95,.05/3,.05/3,.05/3). At publication v>0, set
rho=1/(v+1), retain version v-1, and retire v-2. For normalized old weights w,
retired mass d and the same role prior r:

```text
survivor i:   w'_i = (1-rho) w_i
incoming j:   w'_j = [rho + (1-rho)d] r_j
```

This is a linear stochastic transition: survivor paths stay with probability
1-rho and otherwise enter the new bank; retired paths enter the new bank with
probability1. Do not condition on surviving mass or reinterpret a retired
likelihood as a new snapshot's observation. Fixed share .001 follows each issue
over the active banks, with uniform bank prior and the declared within-bank
role prior. Publication has unit emission, not an extra outcome.

An unresolved issue has unit emission. A received label uses its stored
eight-slot probability vector and generation IDs. Refilter the bounded suffix
from its committed checkpoint through historical publications. A late outcome
can affect the current path posterior through those transitions; it cannot
directly score a new table in a recycled slot. Known/censored prefix messages
are committed exactly once. Scalar messages do not retain retired model tables.

The source's immediate-feedback growing-expert recursion motivates this design;
bounded retirement and delayed refiltering are explicit adaptations, not a claim
to its growing-ensemble regret theorem. Before the first replacement, immediate
feedback agrees with the existing four-role Markov component.

## Detached forecast law

Publication validates and copies the four tables using the existing helper.
Retained banks must have the same normalized input measure; a changed measure
is rejected atomically. A snapshot contains fixed banks, identities and weights.
Its forecast is the corresponding convex mixture of the eight conditional
forecasts. Exhaustive completion averaging verifies coherence over all3^9
partial states in the uniform finite fixture.

Caller mutation cannot rewrite an admitted table. Old previews remain readable
but cannot issue after publication or feedback changes the current law. A failed
table validation or stale issue does not partially commit the filter.

The owner retains at most two model banks. Callers retaining detached previews
can extend old-table lifetimes; this is not a hard bound on all process memory.
The filter itself is a fixed-size value with no model pointers in messages.

## Verification and bug-hunt scope

- Literal path enumeration across three admissions and four share settings,
  with three delayed arrival orders, matches the log-space state recursion.
- Independent probability-space full-history filtering agrees throughout256
  issues, eight publications, reversed releases, missing labels, expiry and
  multiple ring wraps.
- Immediate first-bank limits match fixed Markov at share0,.001,.2,1.
- Tests cover retired-generation feedback, duplicate/future/expired feedback,
  full capacity, stale identity, invalid probabilities, failed publications,
  changed input measures, detached previews and zero-share underflow recovery.
- Focused race suite PASS,2.369s; combined snapshot/Markov/hazard/log component
  race suite PASS,2.615s; vet PASS. No behavioral defect found in this round.
- Reader reentrancy and gate lifecycle are not tested here because this component
  has no reader callback or acceptance gate yet.

## Performance

[All benchmark repeats](../docs/experiments/mmm-snapshot-component-benchmarks.txt),
Apple M4,300ms/count3:

- 256-frame immediate filter lifecycle: .475-.483ms,0 heap allocations.
- 256-frame delay31 lifecycle:5.152-5.226ms,0 heap allocations.
- First detached four-model publication:.733-.744ms,1,261,568 allocated bytes.

Filter issue/admission arithmetic is O(M), delivery/expiry arithmetic O(D*M),
with M<=8 and live suffix D<=80. Transactional value copies additionally touch
the entire capacity C=80: total issue/admission work is O(C*M), delivery/expiry
O((C+D)*M), not O(M) if capacity is treated as a scaling parameter.
Publication table validation/copy additionally costs O(4*3^9).
Heap allocation measurements do not include fixed stack state. No controller,
gate, model fitting or persistence is included; these are not serving p99s or
apples-to-apples comparisons against earlier complete-journal benchmarks.

## Source hashes

```text
bda1617c9194e89e47f8698803fe9a70656c867caf14d70fe294a86f1308b3b5 snapshot_advice.go
ff143fab102f9ff9bceebb033d7b2c8d6edf63c9c18d9feff5a447714150d8a8 snapshot_advice_test.go
4c64b5829115a028154bee6532ffcc1cb3b54af73ee52c57a2fea50bb0eaad0a snapshot_bank.go
d10e81043608830de8f8e77b6bcc70a9c32951796f8fb33a4aa48f0a2af54b13 snapshot_bank_test.go
8616fd91c0d22ae8739062ea02cabbc72c1bd22213829b752ae53f4e4b001bc4 snapshot_advice_benchmark_test.go
```

## Still required

Add identity-scoped acceptance tests and matched total test budgets, then wire
the coherent observation controller and issued-law checks. Rejection must remove
the same predictor from preview and served law; late evidence must retain its
original gate ownership. Test that complete lifecycle before freezing a fresh
quality comparison with all existing stationary and recovery protections.
The broader seven-direction goal remains open, including independent generators,
real prospective tasks and durable loaded serving.
