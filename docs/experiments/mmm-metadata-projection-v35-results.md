# Request-Local Metadata Projection V35 Results

2026-10-03. **All16 normal adoption screens FAIL. No adoption.**
Component integrity survives; original offered latency does not. All seven
whole goals remain OPEN. Production, private corpora and whitepaper unchanged.

## Evidence

1161 source/protocol/auditor files frozen before16new trials, each128writes,
128Recalls and16outcomes. Total2048writes/Recalls,256outcomes and307200candidate
decisions. Native wire, source epochs, exact nomination, as-of, proper-law
arithmetic and full commit-chain/prefix replay checks survive. Forecast error0.
All projected trials capture129private request views and remove getter-owner
acquisitions; native getters/full posterior comparisons still execute.

Frozen race preflight PASS: native differential, copy ownership/native inequality,
held-owner getter, expiry/request/query/vector/selection/old-AsOf rejection,
future versus visible inserts, runtime gap, reopen, cancel/Close, complete
inverse-AST getter/validity/workload checks and31method accounting. Existing
archive retry/lifecycle/interruption checks, vet and core race repeats PASS.

The prospectively frozen auditor verifies all16rows, then FAILS its trace-key
control: it targets the setup prime whose complete nomination was not saved.
Original source and failure log are preserved. A **post-hoc supplement**, not
a prospective success, changes only that corruption target to a delivered
request and passes23corruptions/all16rows. Prime full nomination and independently
bounded acknowledgment remain unverified; all128delivered Recalls per trial
have saved nomination and ack checks. Supplement setup initially omitted VM
Buffer; that failure is separately preserved. No normal run was repeated.

Authority: `research/metadata-projection-v35/{freeze,run,supplement}.json`,
`raw.ndjson`, `audit-attempt.log`, `technical-prefreeze/*`, setup-failure notes.
Raw SHA256 `2ad2484d5cbe9d258054a672ff0924493f1bb00b3deb37f25f51381a2f92a6bf`.
Freeze SHA256 `6e2dfae3132d90a369dce4ae3c61bd59f7100c3f0c1bd162077a37a0dbc16b99`.
Apple M4/10logical CPUs/16GiB. Collector exit0 means completed, not adoption.

## Loaded Results

Off/on is request-local projection. Joined=scheduled joint witness;
archive=bounded archive with shared ancestry and deferred audit-proof formatting.
Fixed order/two reps constrain attribution; not a significance claim.

| Rep | Visibility | Storage | Projection | Call p99 ms | Offer p99 ms | Write p99 ms | Outcome p99/max ms |
|---:|---|---|---|---:|---:|---:|---:|
|1|Future|Joined|Off|58.848|191.306|88.400|402.484|
|1|Future|Joined|On|66.505|180.729|96.966|396.404|
|1|Future|Archive|Off|75.430|257.801|98.483|435.149|
|1|Future|Archive|On|88.125|296.950|107.424|421.247|
|1|Visible|Joined|Off|57.416|161.132|85.893|377.986|
|1|Visible|Joined|On|60.372|189.923|93.250|396.687|
|1|Visible|Archive|Off|83.070|274.948|111.934|470.091|
|1|Visible|Archive|On|89.945|290.431|108.079|462.970|
|2|Future|Joined|Off|56.524|192.436|92.248|400.204|
|2|Future|Joined|On|61.115|215.195|99.791|429.890|
|2|Future|Archive|Off|85.447|286.366|105.277|467.612|
|2|Future|Archive|On|82.999|311.718|110.313|484.336|
|2|Visible|Joined|Off|54.696|153.735|81.852|363.978|
|2|Visible|Joined|On|61.609|190.917|91.584|398.225|
|2|Visible|Archive|Off|68.811|252.057|99.525|463.843|
|2|Visible|Archive|On|84.313|287.548|109.879|486.483|

Every row passes call-only100ms and write/view250ms, but FAILS offered
Recall100ms and outcome100/250ms. Seven of eight paired offered tails worsen;
one future joined cell improves but still fails. Per-request JSON copying totals
4159752-4411102bytes per projected trial (4.16-4.41MB decimal), O(log+sources)
per request, not O(1). Future trials invoke19608metadata getters; visible
trials1008-1158. Getter success counters independently reconstructed from traces.

All16future outcome sources eventually influence forecasts in both arms;
maximum first-use408.718-504.799ms projected versus413.797-491.326ms controls.
Visible transport0 as required. Eventual uses reflect backlog opportunities,
not useful-before-staleness improvement. No agent-answer quality evidence.

## Next Lead

Cache an immutable read core at a validated runtime head, preserving native
reads and external equality. Journal-only publications cannot change the core
after initial certificate binding; initial nil-certificate state must not be
cached across that transition. Prove this exact boundary and reject gaps,
foreign/expired requests and source changes before fresh load trials. No borrowed
mutable maps, permanent epoch retag or removed FULL durability. Continue
untouched-domain retrieval and richer adaptive models alongside serving.
