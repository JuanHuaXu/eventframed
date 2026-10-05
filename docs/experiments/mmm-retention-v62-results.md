# V62: shared bank and actual scored-law integration

## Verdict and scope

[Protocol](mmm-retention-v62-protocol.md) completed: race, vet, allocation and
three-repeat benchmarks, all terminal exit 0. The retention selector is now
wired to the actual three-window model, not an isolated posterior that cannot
affect scoring. Mathematical/lifecycle and nominal150-member memory components
PASS. The200-member constructor still FAILS8MiB. No new broad accuracy/recovery
result, adaptive observer comparison, untouched agent result, loaded serving
or whole-goal success is claimed. All seven WHOLE goals remain OPEN/ACTIVE.

## What changed

New research-only bank shares immutable copied baselines, ranks, rates, priors
and likelihood tables across600/1,200/2,400issued-position windows. All mutable
model counts, posterior/support state and journals remain separate in V62.
The next-trial(Y,W1,W2)joint is computed from each actual tree/noise posterior;
the coherent adapter derives its clean/measurement marginals. Member-specific
selector weights mix those forecasts before issue, producing the scored output.

Immutable issue-time kernels grade first/paired observations at their original
ordinals. Later W2 requests smooth the OLD issued kernel; a new-trial forecast
cannot silently replace that original law. Old grading may move selection
weights after a short model expires that observation, but does not restore the
expired model factor. All dependent corrections normalize before any commit.
Atomicity is serial in-memory API atomicity, not concurrent/durable publication.

Twenty-three functional race-test roots pass. Independent top-down tree/noise
reconstruction and frozen-selector grading reproduce25,758bank forecasts across
three depths and three delay schedules, including reordered/missing replies and
distinct expiry. Independent full-ledger reconstruction fences pass. The inherited
4,656forecast/120value checks and2,000explicit expert-path checks also pass.
Ownership tests distinguish shared immutable tables from separate mutable states.
Four injected normalization faults verify all-or-nothing publication, and a
later-child issue-cap failure cannot prematurely expire another child's suffix.
Future-outcome forks agree until receipt and differ non-vacuously after receipt.

## Cost

Actual allocated bytes, measured with MemStats; includes constructor work and
retains first-run/background allocation overhead. These are NOT RSS values.

| Capacity | Shared bank | Unshared bank |8MiBcap|
| --- | ---: | ---: | --- |
|150members|7,951,176|10,001,168|Shared PASS; unshared FAIL|
|200members|9,988,752|12,735,888|Both FAIL|

The cap is8,388,608bytes. The nominal150-member bank has437,432bytes of measured
headroom; this is not a claim that additional observer/persistence state fits.
The200-member bank remains1,600,144bytes over cap. No cap was relaxed.

Apple M4,Go1.27.1darwin/arm64, three100ms repetitions:

| Operation | Time | Allocations |
| --- | ---: | ---: |
|150-member shared constructor|2.902-2.912ms|26;~7.946MBsteady allocation|
|200-member shared constructor|3.643-3.661ms|26;~9.989MB|
|Actual mixed clean prediction|815.4-857.4ns|0|
|2,400-frame bank core loop|167.198-171.898ms|26;7,945,744bytes|

The core loop includes construction, all three model issue/expiry/updates,
selector updates,400paired requests/replies and16all-member snapshots. Packets
are nominated by a fixed every-sixth-frame schedule, not a learned observer.
It excludes adaptive candidate scoring, delayed-queue scheduling, source I/O,
persistence, RPC, loaded serving and background freshness. Passing this core
timing does NOT prove the whole400ms experiment gate or sub100ms serving tails.

## Audit trail

[Freeze](../../research/retention-v62-integration/freeze.json):27compiler inputs
plus12supports,39frozen/copy-verifiedfiles; generated test mains separately
recorded and copied. [Completion](../../research/retention-v62-integration/completed.json)
records all four commands and output hashes. Previous checkpoint's166saved
files and generated input were verified before running;14preexisting tracked
hashes stay unchanged. Production/private/paper/publication untouched.

Two pre-freeze repairs are retained. Mechanical constructor-name substitution
also renamed18qualified errors.New calls; inspection restored the imported API
before tests. A non-vacuity fixture initially graded the first-ever trial, whose
three prior kernels are necessarily identical. It failed correctly. The repaired
fixture warms distinct suffixes first, preserves strict non-vacuity and checks
that old grading cannot revive short-window state. [Failed source and metadata](../../research/retention-v62-preflight/failure.json)
are preserved; no inference rule or scientific threshold changed.

## Next

The remaining memory gap motivates a bank-owned canonical OBSERVATION journal,
not shared posterior state. It requires explicit child-mutation authority fences
and joint/lifecycle equivalence. [V63 direction](../../research/retention-v63-journal-direction.md)
keeps that proposal separate from this frozen result. Same-mixture acquisition,
full controlled quality/recovery/cost tests, broader independent data, useful
certified splits, untouched tasks and loaded freshness remain required under
the original seven-goal objective. No sealed labels or confirmation seed opened.
