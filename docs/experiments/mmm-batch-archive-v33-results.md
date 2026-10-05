# Bounded Archive V33 Results

2026-10-03. Eight NEW normal mixed-load trials complete. **All eight adoption
screens FAIL; archived arm regresses loaded tails in all four matched cells.**
Do not adopt this prototype. Technical/law successes are not goal6 completion.

## Frozen Workload

1151source/protocol/auditor hashes; independently inverse-AST checked V26 offered
rates, producers, selection, full metrics and gates. Both arms joined+scheduled;
archive-only factor shortens exclusion and adds bounded capture/ancestry work.
Per trial128writes/128Recalls and16mixed outcomes all acknowledge:1024/1024/128
total,153600candidate decisions, eight prime journals. Exactly150nominees/request,
all original pins/laws/source epochs/wires and144mutations per trial retained.
Synthetic public vector fixtures only, no production/private-session corpus.

| Rep | Mode | Arm | Call p99 ms | Offer p99 ms | Write p99 ms | Outcome p99/max ms | View Max ms |
|---:|---|---|---:|---:|---:|---:|---:|
|1|Future|Control|54.742|183.567|88.201|390.377|38.505|
|1|Future|Archive|79.874|294.728|116.515|475.174|52.594|
|1|Visible|Control|68.960|196.794|105.103|407.130|42.765|
|1|Visible|Archive|92.242|293.280|113.724|470.699|53.444|
|2|Future|Control|57.051|147.555|86.964|366.208|36.191|
|2|Future|Archive|83.500|304.175|115.937|491.284|56.249|
|2|Visible|Control|52.006|148.632|83.706|360.725|35.867|
|2|Visible|Archive|72.140|271.063|105.537|451.187|52.447|

Original Recall call/offer100ms, outcome100/250ms, write250ms and view250ms gates
unchanged. Call-only passes every row, but offered Recall and outcomes fail every
row. A passing Go collector means the harness completed, NOT adoption success.
No queue work hidden: archive accepted/finished129 per trial, rejected0, peak8
under cap128, active0 after drain.129bindings use40/43/42/41archive batches.

Future control learned1023/1022decisions, cross-epoch884/896; archive1134/1019,
cross918/825. All16sources eventually observed in each future arm; first-use
max403.012/379.300ms control versus494.433/520.751ms archive. Longer backlog can
inflate eventual uses; no freshness/quality improvement. Visible all0transport.
Independent ordinary Beta predictive arithmetic max discrepancy0.

## Integrity And Technical Tests

Preflight frozen1147files: race cap3 rejects fourth; cancellation and Close wait
for accepted durability; original/same-wire retry, forged stored root, reopened
service and three interruption points PASS. Sealed V31 sequential retry fails
as expected after journal-only publication; V33 retains the first root and checks
its original ancestry instead of replacing it. Vet and three core race repeats
PASS. Separate race-enabled workload/native inverse-AST checks PASS.

Original frozen auditor checks all8rows, then fails its premature-ack corruption
control: it corrupts the PRIME batch, which has no corresponding delivered read
timestamp in raw. Preserve exact failed attempt log and original source. A
post-hoc supplementary auditor uses the SAME row checker and targets a delivered
batch; all16corruptions rejected. All128loaded read acknowledgments are bounded;
unrecorded setup-prime ack cannot be independently time-bounded from this raw.
No runtime/raw/gate rewrite or rerun of the confirmation to repair the checker.

Raw SHA256: `16476e5d63fd40580b7091c5a2123fbe1747f01abc3e6147531820c323751d85`.
Evidence: `research/batch-archive-v33/{freeze,run,supplement}.json`, `raw.ndjson`,
`load.log`, `audit-attempt.log`, `technical-prefreeze/*`. No `audit.json` success
artifact is asserted for the failed frozen auditor.

## Attribution, Not Confirmation

Separate sealed-workload profile rerun completed, with CPU/mutex/block artifacts,
binary, raw and hashes. Instrumented timings are NOT another normal confirmation.
Aggregate mutex waiter delay6.09s;4.624s attributed to archival apply's deferred
owner unlock versus~79.7ms joined apply. This is aggregate waiter time, NOT a
per-request wall delay or proof that owner scheduling alone explains regression.
Block trace outcome owner wait273.33ms aggregate across the diagnostic run.

Focused sampled CPU: archive apply~0.42s cumulative, native archive append~0.31s,
ancestry validation~0.11s, capture~0.16s. Cumulative stacks overlap; do not add them
or call ancestry the sole bottleneck. Whole-profile runtime/HNSW setup samples
are not a serving-only CPU breakdown. These limited profiles support testing
contention-aware publication/projection and amortized validation, not bypassing
native readback or declaring a measured rescue.

Archive captures also serialize full encoded proof records before job ack for
this audit; that instrumentation cost is included in its measured tails. Control
does not generate the extra capture proof. A future fair decomposition should
retain immutable captures but materialize benchmark-only proof strings AFTER
serving drains, or charge identical proof logging to both arms. Do not silently
remove this cost and relabel V33 as a faster result.

## Next Leads

1. Derive an immutable dependency projection with complete mutation coverage,
   so ordinary metadata reads need not repeatedly acquire the publication owner.
   Verify native equality/as-of/source epochs before any loaded adoption claim.
2. Validate one complete ancestry pass per batch with per-capture membership,
   not repeated full scans; preserve all wire/capture/unknown-history controls.
3. Separate publication scheduling from active read dependencies without priority
   inversion; measure original offered tails and freshness on a NEW frozen cohort.
4. Keep untouched-domain retrieval and broader adaptive experiments in scope;
   this negative serving result does not settle quality goals1-5/7.

All seven whole goals OPEN. Production/private corpora/whitepaper unchanged.
