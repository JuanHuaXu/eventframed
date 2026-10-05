# V62 shared bank and scored-law integration protocol

ISOLATED integration/component study, not adoption. All seven WHOLE goals stay
OPEN/ACTIVE. Previous V60 quality failure is preserved. V61's selector was not
connected to actual model kernels and an unshared 150-member bank exceeds 8 MiB.

## Pre-patch classification and invariant

Duplicated immutable tables are a CONFIRMED allocation cause, not the unique
cause of accuracy or latency failures. Alternative causes include duplicated
mutable journals, observer computation and stale-window/model mismatch. This
patch targets only the first and the unconnected-law gap. It does not assume
sharing improves forecasts or that a correct selector rescues practical learning.
No upstream production patch is sought: all edits are new research-only paths.

Three V60 models retain600/1,200/2,400 ISSUED positions. Share package-private
baseline copies, feature ranks, rates, priors and likelihood tables; each owns
all counts, support counters, posteriors, pending/epoch state and journals.
No caller-visible writable template is exposed. Siblings must reproduce the
corresponding unshared model at every visible-evidence boundary.

Extract the actual next-trial joint law on (Y,W1,W2) from each window's current
tree/noise posterior. Derive clean Y and observed W1/W2 marginals through the
joint-law adapter. The member-specific selector's next weights mix these actual
clean forecasts BEFORE issue; that mixture is the issued scored output. Store
each issue's joint kernel for original-position first/paired grading. Later
W2 requests use the selector's smoothed issued kernel, NOT a silently replaced
new-trial law. A new trial kernel must never be substituted for old Y evidence.

Prepare/normalize every window and selector correction before publishing any
dependent update. Commit only after all pass; a failed late child cannot leave
an earlier child updated. This is serial API atomicity, NOT concurrent-serving
or durable-publication atomicity. Epoch reset invalidates all old tickets.

Old grading may revise selector weights even when a short model has expired
that outcome, but cannot revive expired model factors. This is a separate ledger
lifecycle, not extra independent evidence. Missing replies remain explicit.

## Verification and cost

Before timing, run model/selector/reference and integration race tests, including
independent top-down kernel reconstruction and frozen-selector comparison.
Exercise multiple depths, delayed/reordered/missing evidence, old grading,
zero support, caps, immutable/mutable ownership and failure at each child or
selector. Fork future outcomes at the receipt boundary with identical prior
forecasts and a non-vacuous post-receipt difference. Keep mathematical tests
separate from quality experiments and preserve any failed preflight fixture.

Freeze compiler/test closure, all new sources and 14 preexisting tracked hashes.
Run race/vet/allocation/microbench serially with complete command logs and exit
codes. Measure shared AND unshared allocation at150 AND200members. The unchanged
nominal150-member8MiB gate is evaluated from actual allocation, not estimates;
report200-member failures rather than discarding them. Allocated bytes are not RSS.

Three-repeat100ms benchmarks cover constructors, actual mixed prediction and a
2,400-frame core loop with400 paired packets and16all-member snapshots. Count
all three models, selector and update work. This fixed packet schedule has no
adaptive candidate integration, delayed queues, persistence or loaded serving:
it cannot prove the whole400ms experiment gate or Goal6freshness/sub100ms tails.

## Still required

Same-mixture acquisition with a defined objective and uncertainty/cost accounting;
complete controlled quality/recovery experiments including every changing/noisy
regime and independent replay; broader untouched seeds; useful error-controlled
Anti-Pigeon outcomes; untouched labeled agent utility; loaded serving/freshness;
equal TOTAL-cost observation superiority. No narrower component success counts
as one of the seven completed goals. No new confirmation or private data is used.
