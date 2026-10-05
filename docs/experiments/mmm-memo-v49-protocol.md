# V49 exact constructor memoization

V48's audited diagnostic retains its scientific FAIL: maximum loop422.620583ms
against the unchanged400ms requirement. A repeated constructor prior projection
is a performance hypothesis, not a proven explanation for every slow cell.
Other causes include delayed replay, extra mixer layers, and host variability.
No mathematical quality policy, threshold, prior, evidence or gate changes.

New research-only constructors reuse identical finite prior arrays within one
invocation. Fixed128-entry FIFO, exact float64 mean key, fixed strength/mode
per invocation. No quantization, cross-constructor state, observation-dependent
cache or approximate inference. On miss/eviction compute the ORIGINAL prior.
Arrays copied by value. Original constructors and all runtime methods untouched.
Epoch rebuild uses the original constructor: no reset-performance claim.
Memory bound adds fixed128*(one mean+21masses), not a corpus-sized cache.

Require exact original/candidate state equality over narrow/rich families,
density/moment priors, boundary/mixed means, strengths/hazards, shared/private
updates, delayed/canceled outcomes, invalid contracts, input mutation and epoch
replacement. Reject on ANY bit difference. Audit all84existing V48 diagnostic
cells/252arms against their independently validated original advice, weights,
served forecasts and receipts. This is consumed-data computational ablation,
NOT new quality confirmation or policy selection. Compare paired measured cost
with original; preserve misses. Constructor<=8MiB, complete loop<=400ms remain.
Output buffers precede timer as in V48; generation/audit/scoring/serialization
are separately recorded. No loaded-serving guarantee follows.

Freeze exact source closure before experiment, execute unit/race/vet plus cost
and full ablation, store terminal commands/hashes/raw timing and independent
readback. Normal V48 cohorts remain unused until actually executed. All seven
WHOLE goals remain OPEN. Production/private data/whitepaper/publishing untouched.
