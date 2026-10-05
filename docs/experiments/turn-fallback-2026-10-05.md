# Turn Fallback Audit, 2026-10-05

Current status: semantic validation passes for frozen final turn.go SHA256
`88527b6032ebdd1f9d50102134a16fbcddabbf093e0aed708b5e76644fe51cb3`.
Both approved extraction bugs are repaired. Full frame functional/race tests and
vet pass after the parent fixed the stale multiline-quote test expectation.
The capture invariant and regression assertions remain intact. Pre-fix and
intermediate-candidate measurements are preserved separately below; the parent
owns the final-source benchmark in runner 45416.

The approved repair owns only turn.go, query.go, this audit test, and this report.
No deployment, Git writes, or edits to other files are authorized.

## Pre-Fix Audit (Preserved)

Everything in this section describes the earlier read-only phase. Its "current"
measurement labels mean the pre-fix turn.go SHA256 `88f69778c85eebd7139ff292c255741381b28f3b6796ae00966c89eb4a48c401`,
not the repaired candidate. All original 360 measurement tuples are retained.

Status: audit and matched benchmark complete. Not ready to approve publication;
confirmed semantic failures remain. No production changes were made.

## Boundary

Only this new report and `internal/frame/turn_fallback_audit_test.go` are owned by
this audit. Existing production, plugin, instruction, and research files are
read-only. No checkout replacement, branch/index edits, commits, pushes,
installs, deployments, or unrelated scratch cleanup are authorized.

Target: `/Volumes/Data/eventframed/internal/frame/turn.go` against HEAD
`1a7edb62b6be4031fd01ebeab8071b17303a7815`.
Current SHA256: `88f69778c85eebd7139ff292c255741381b28f3b6796ae00966c89eb4a48c401`.

## Exact Current Diff From HEAD

The complete turn.go diff is 7 added and 5 removed lines, with no staged edits.

```diff
diff --git a/internal/frame/turn.go b/internal/frame/turn.go
index d73e55b..d376be3 100644
--- a/internal/frame/turn.go
+++ b/internal/frame/turn.go
@@ -92,8 +92,9 @@ func FromTurn(turn model.TurnCapture) model.Event {
 
 func firstField(sources []sourceText, patterns []pattern, confidence float64, accept func(string) bool, fallback model.Field) model.Field {
 	for _, source := range sources {
+		clean := unquoted(source.text)
 		for _, candidate := range patterns {
-			indices := candidate.expression.FindStringSubmatchIndex(source.text)
+			indices := candidate.expression.FindStringSubmatchIndex(clean)
 			group := candidate.group * 2
 			if len(indices) <= group+1 || indices[group] < 0 {
 				continue
@@ -113,14 +114,15 @@ func firstField(sources []sourceText, patterns []pattern, confidence float64, ac
 }
 
 func participantFallback(turn model.TurnCapture, user sourceText) model.Field {
-	if indices := collectivePattern.FindStringIndex(user.text); indices != nil {
-		return field("user and agent", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
+	clean := unquoted(user.text)
+	if indices := collectivePattern.FindStringIndex(clean); indices != nil {
+		return field("unresolved group", model.SourceInferred, 0, span(user.name, indices[0], indices[1]))
 	}
-	if indices := firstPersonPattern.FindStringIndex(user.text); indices != nil {
+	if indices := firstPersonPattern.FindStringIndex(clean); indices != nil {
 		return field("user", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
 	}
 	if turn.AgentID != "" {
-		if indices := secondPersonPattern.FindStringIndex(user.text); indices != nil {
+		if indices := secondPersonPattern.FindStringIndex(clean); indices != nil {
 			return field("agent:"+turn.AgentID, model.SourceInferred, .70, span(user.name, indices[0], indices[1]))
 		}
 		return field("user and agent:"+turn.AgentID, model.SourceInferred, .55, "turn participants")
```

## Matched Reconstruction And Fixtures

The baseline copies the full HEAD `FromTurn`, `firstField`, and
`participantFallback` into test-only functions, with only function names/calls
renamed. AST comparisons verify those bodies against HEAD and check all shared
turn.go declarations. The current file hash is pinned. HEAD has no quote mask
in these paths: adding today's mask to the baseline would not measure the
actual patch. Other working-tree changes are held constant, not rolled back.

Fixtures are deterministic, newly authored public synthetic text; no private
chat, user card, network data, or production corpus is read. Both implementations
receive identical complete TurnCapture inputs, including metadata and retrieved
IDs. Benchmarks time the full FromTurn result, retain it in the same sink, report
allocations, and exclude fixture construction and source verification.

Five profiles (early match, late match, no match, quoted claims, collective
fallback) use exactly 256 B, 2 KiB, or 16 KiB per role. Payload bytes/op are the
sum of user and assistant text sizes. Forward and reverse implementation order
are run separately with repeated timings. Latency is ns/op, lower is better;
throughput is its inverse, higher is better. Percentage throughput changes must
not be reported as percentage latency changes.

## Patch Reasoning Gate

Loaded `patch-reasoning-audit` before test edits. The invariant under review is:
quoted content is not direct actor/claim evidence, and a collective reference
does not establish a specific user-agent group. Raw content and first-statement
summaries remain retained metadata. Explicit adapter identities remain separate
from names claimed in prose. The lifecycle is post-contract deterministic
extraction, not retrieval/provider delivery; no provider-visible claim is made.

Candidate explanations to distinguish: fallback guesses a group; earlier name
extraction bypasses fallback; masked regex spans reintroduce quoted payload when
sliced from raw text. Tests sample user/assistant roles, quote forms, live text
after quotes, ambiguous groups, adapter identities, and unchanged controls.
Falsifiers are zero-confidence collective output from the full FromTurn path,
and fallback rather than observed fields for quoted-only clause payloads.

Shared firstField also affects FromText and QueryText. FromTurnWithIdentities
starts with FromTurn but replaces Who using its own conservative resolver;
Where/When/Why/How can retain base extractor results. The local identity.go is
untracked and absent at HEAD, so no claim of an end-to-end HEAD identity/service
benchmark is possible from this matched reconstruction. No upstream fix or
network search is needed for this local read-only audit; no fix is authored.

Hot path: participant fallback eagerly runs even when Who has an explicit name.
Each firstField masks once per visited source, repeated across five field calls;
full misses visit both roles (ten masks), plus the fallback user mask. Masking
scans and allocates proportional to input length, preserves byte positions, and
has no I/O. Benchmarks distinguish early exits from late matches/full misses.

## Confirmed Findings

Classification below separates current regressions from inherited limitations.
No finding authorizes edits outside the two owned files.

### [P2] Masked Captures Can Yield A Lone Quote Or Restore Quoted Payload

**confirmed**. In [turn.go](/Volumes/Data/eventframed/internal/frame/turn.go:97),
the regex runs on masked text, but the capture is sliced from the original text
at line 103 and checked for emptiness in that original representation. The
reason pattern accepts whitespace as its entire payload.

User input `Please deploy because "cache failed".` produces:

| Version | Why value | Source | Confidence | Evidence |
| --- | --- | --- | ---: | --- |
| HEAD | `"cache failed"` | observed | 0.90 | `user[bytes:22:36]` |
| Current | `"` (one closing quote) | observed | 0.90 | `user[bytes:35:36]` |

The mask is `Please deploy because               .`. Greedy separator
whitespace leaves one space for the nonempty reason capture; slicing its raw
byte range returns a delimiter. This lone-delimiter field is an actual patch
regression, not merely an intended change in quote policy.

The same mask/raw-span mismatch also leaves `using \`console\`` and
`after "approval"` as directly extracted How and When values. Those two
outputs already occur at HEAD: they demonstrate incomplete exclusion of quoted
payload, not new regressions. The tests enforce fallback when the entire clause
argument is masked, not removal of all quotes from retained metadata.

Both FromTurn and FromTurnWithIdentities reproduce all three clause cases,
for both user and assistant roles. Assistant values retain synthetic provenance
and the existing 0.12 confidence reduction, but remain direct extractions.
Identity-aware reference resolution masks quoted pronouns; it does not repair
these base field values. Shared firstField means the new delimiter result also
affects FromText/QueryText by the same call path; no service/provider behavior
is inferred from that static reachability.

Regression: `TestTurnFallbackAuditKnownBugRegressions/quoted-only-*`.
Positive control: wholly quoted claims fall back, while a live unquoted claim
after a multibyte quote keeps the correct original-byte evidence.

### [P2] Capitalized We Bypasses Collective Abstention

**confirmed, pre-existing at HEAD**. In
[turn.go](/Volumes/Data/eventframed/internal/frame/turn.go:38), the capitalized
name pattern accepts `We`; the extracted match wins over the eagerly computed
participant fallback at line 81.

`We will deploy.` yields `Who={Value:We, Source:observed, Confidence:0.88,
Evidence:user[bytes:0:2]}` on both HEAD and current. Lowercase `we plan a
release.` and uppercase `We plan a release.` do reach the new zero-confidence
fallback. Thus the fallback correction does not cover the full actor path.

FromTurnWithIdentities correctly abstains on all three variants, even with an
explicit speaker card, a participant roster, and a previous named actor. It
replaces base Who and records ambiguous-reference metadata; adapter `I` and
`You` subjects still resolve to the supplied speaker and agent. Self-claimed
`I am Another Person.` also retains the adapter/card speaker identity.

Regression: `TestTurnFallbackAuditKnownBugRegressions/capitalized-collective-bypasses-fallback`.
This is not evidence that the current patch introduced the actor bug.

### Follow-Up Classification

**recommendation only**: the owner should decide the admission rule for captures
whose payload is wholly masked, and address pronouns accepted by the name path
if full FromTurn abstention is a requirement. Mask reuse across fields is also
a possible optimization, subject to preserving source order, offsets, synthetic
confidence, and all sibling extraction APIs. No production fix, general policy
change, or cleanup was made. Upstream-related-fix investigation remains deferred
until an implementation/publication task is authorized.

## Validation

Final controls cover collective fallback precedence, adapter identity, quoted
claims/pronouns in eight quote forms, Unicode byte alignment, contractions,
synthetic evidence, raw metadata, full-event parity on unchanged inputs, and
AST reconstruction fidelity.

| Check | Result |
| --- | --- |
| Audit positive controls, count=3 | PASS, 0.469 s |
| Audit positive controls, race, count=1 | PASS, 2.448 s |
| Desired-contract bug regressions | FAIL: four groups, seven leaf inputs |
| Full internal/frame package | FAIL only the active audit bug regressions, 0.249 s |
| Full internal/frame package, race | Same semantic failures, 1.474 s; no data-race reports |
| go vet -mod=readonly ./internal/frame | PASS |
| gofmt -d on the new test file | PASS, no formatting differences |
| Final standalone baseline-fidelity check | PASS, 0.193 s |
| Clean forward matched benchmark | PASS, 44.114 s, 180 readings |
| Clean reverse matched benchmark | PASS, 43.716 s, 180 readings |

The failing tests deliberately assert the desired semantics; they are neither
skipped nor inverted into expected-bug assertions. Ordinary package tests will
remain red while these audit regressions are present and unresolved. Race
instrumentation ran, but the full race command is not described as passing.

Commands (run from `/Volumes/Data/eventframed`):

```sh
go test -mod=readonly ./internal/frame -run '^TestTurnFallbackAudit' -count=1
go test -mod=readonly ./internal/frame -count=1
go test -mod=readonly -race ./internal/frame -count=1

go test -mod=readonly ./internal/frame \
  -run '^TestTurnFallbackAudit(BaselineFidelity|ParticipantControls|IdentityControls|QuotedClaims|QuotedPronouns|QuoteMaskAlignment|UnquotedEvidenceAfterQuotes|MatchedControls)$' -count=3
go test -mod=readonly -race ./internal/frame \
  -run '^TestTurnFallbackAudit(BaselineFidelity|ParticipantControls|IdentityControls|QuotedClaims|QuotedPronouns|QuoteMaskAlignment|UnquotedEvidenceAfterQuotes|MatchedControls)$' -count=1

env GOMAXPROCS=1 go test -mod=readonly ./internal/frame -run '^$' \
  -bench '^BenchmarkTurnFallbackAuditMatched$' -benchmem -benchtime=200ms -count=6 -cpu=1
env GOMAXPROCS=1 go test -mod=readonly ./internal/frame -run '^$' \
  -bench '^BenchmarkTurnFallbackAuditMatchedReverse$' -benchmem -benchtime=200ms -count=6 -cpu=1
```

## Benchmark Results

Host: Apple M4, darwin/arm64, Go 1.27.1, 10 physical/logical cores, 16 GiB RAM.
Measurements use GOMAXPROCS=1 and one serial benchmark worker. There is no CPU
affinity, exclusive host reservation, or assertion that unrelated processes
were idle. Both accepted processes ran after test/race completion, with no
concurrent audit workload. An initial benchmark overlapped briefly with race
compilation and was excluded entirely, not selectively pruned.

Each implementation has six readings in forward order and six in reverse
order per profile/size, 360 total accepted readings. The two benchmark commands
run the implementations in opposite orders; Go groups count repetitions by
subbenchmark. These are repeated matched inputs, not independently randomized
or one-for-one interleaved paired trials.

The table reports medians of all 12 readings per implementation. Percentage
columns are calculated from those median ns/op values, not rounded MB/s.
B/op and allocs/op are likewise medians; all original readings are retained
below. A fractional median is the average of the middle two reported values.

| Profile | Per role | HEAD ns/op | Current ns/op | Latency change | Throughput change | HEAD -> current B/op | HEAD -> current allocs/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| early | 256 B | 35912 | 40550 | +12.9% | -11.4% | 4368 -> 9457 | 70 -> 94 |
| early | 2 KiB | 205567 | 241172 | +17.3% | -14.8% | 8657 -> 49368 | 70 -> 94 |
| early | 16 KiB | 1727915.5 | 2182420.5 | +26.3% | -20.8% | 44763 -> 368928 | 70 -> 94 |
| late | 256 B | 56375 | 61035.5 | +8.3% | -7.6% | 3928 -> 9112 | 70 -> 94 |
| late | 2 KiB | 448830 | 500760 | +11.6% | -10.4% | 8297 -> 49007 | 80 -> 104 |
| late | 16 KiB | 3663451 | 4330373 | +18.2% | -15.4% | 44401 -> 368556 | 80 -> 104 |
| no-match | 256 B | 142169.5 | 151199 | +6.4% | -6.0% | 2216 -> 11721 | 41 -> 85 |
| no-match | 2 KiB | 1153836.5 | 1291605 | +11.9% | -10.7% | 6506 -> 81139 | 41 -> 85 |
| no-match | 16 KiB | 9746804.5 | 10996241 | +12.8% | -11.4% | 42623 -> 636899.5 | 41 -> 86 |
| quoted | 256 B | 55238 | 133415 | +141.5% | -58.6% | 2808 -> 11589 | 69 -> 96 |
| quoted | 2 KiB | 451936 | 1277166.5 | +182.6% | -64.6% | 7177 -> 81183 | 79 -> 96 |
| quoted | 16 KiB | 3735728.5 | 11176880 | +199.2% | -66.6% | 43281 -> 636950 | 79 -> 97 |
| collective | 256 B | 136212.5 | 146467.5 | +7.5% | -7.0% | 2312 -> 11801 | 42 -> 86 |
| collective | 2 KiB | 1089477.5 | 1240212.5 | +13.8% | -12.2% | 6602 -> 81220 | 42 -> 86 |
| collective | 16 KiB | 9003544 | 10558844 | +17.3% | -14.7% | 42717 -> 636974 | 42 -> 87 |

For fixed payload size, define `r = current_ns / HEAD_ns`.
Latency change is `100 * (r - 1)`; throughput change is
`100 * (1/r - 1)`. For example, quoted 16 KiB/role latency is about +199.2%
(2.99x), while throughput is -66.6%, not -199.2%.

### Timing Spread And Order Check

| Profile | Per role | HEAD ns/op min-max | Current ns/op min-max | Forward latency change | Reverse latency change |
| --- | --- | ---: | ---: | ---: | ---: |
| early | 256 B | 34622-36146 | 39622-49418 | +18.9% | +12.2% |
| early | 2 KiB | 201025-207569 | 236276-246018 | +20.2% | +16.8% |
| early | 16 KiB | 1604047-1879643 | 2017988-2296064 | +24.9% | +29.9% |
| late | 256 B | 55162-58393 | 59669-61834 | +8.2% | +7.8% |
| late | 2 KiB | 437389-473046 | 484108-524816 | +9.5% | +13.9% |
| late | 16 KiB | 3591182-3764527 | 4175028-4380743 | +19.4% | +16.8% |
| no-match | 256 B | 139553-145972 | 147610-152894 | +6.7% | +6.4% |
| no-match | 2 KiB | 1137164-1165205 | 1253155-1314837 | +13.6% | +10.8% |
| no-match | 16 KiB | 9430888-11708017 | 10814799-11544017 | +14.6% | +13.8% |
| quoted | 256 B | 54103-64662 | 129036-138164 | +141.7% | +144.6% |
| quoted | 2 KiB | 439267-482985 | 1223450-1360729 | +180.0% | +181.7% |
| quoted | 16 KiB | 3619686-3905581 | 10833491-12027090 | +202.3% | +195.8% |
| collective | 256 B | 130653-166416 | 144568-180393 | +10.2% | +6.7% |
| collective | 2 KiB | 1051759-1108706 | 1220884-1257495 | +17.1% | +13.0% |
| collective | 16 KiB | 8736173-9396167 | 10126205-10935203 | +16.6% | +14.3% |

### Interpretation And Confidence

Unquoted profiles have approximately +6.4% to +26.3% median latency in this
run, equivalent to about -6.0% to -20.8% throughput. Positive parity tests show
identical full events for early, late, and no-match inputs; collective parity
differs only in the intended Who fallback. Added work is consistent with
repeated quote masking even where no quote is present.

Quoted profiles have approximately +141.5% to +199.2% latency and -58.6% to
-66.6% throughput. These are **not equal-output controls**: HEAD finds the
quoted claims, while current excludes them and continues scanning additional
patterns/sources. Their delta is the combined cost of changed extraction paths
and masking, not an isolated estimate of mask overhead.

At 16 KiB/role, early/late allocation medians grow by about 324 KB/op and full
miss/collective cases by about 594 KB/op. Their medians add 24 and 45 allocations,
respectively. Different quoted outcomes also change allocation counts. Some
samples report additional bytes/allocations; runtime/regexp warm-up or pooling
effects are plausible, unprofiled explanations. Those samples are retained
without trimming.

Confidence is high that the source-matched treatments and semantic counterexamples
are reproducible in this checkout. Confidence is moderate in these host-specific
timing ratios, and low in extrapolation to production latency, throughput, or
end-to-end retrieval. Order reversal preserves the direction for every cohort,
but ranges include noise/outliers and no inferential confidence interval or
benchstat significance claim is made. No profiler isolates individual masking,
regex, GC, or allocation contributions; precise causal attribution among those
costs would overstate this evidence.

Sizes are plausible synthetic text envelopes, not a measured production size
distribution. The ASCII neutral filler and five controlled shapes are not a
representative conversational corpus. No private chats were used.

## Final Boundary Check

Only the two new authorized paths were written. No production/plugin edits,
Git writes, commits, pushes, installations, deployments, or scratch deletions
were performed. Tracked status is unchanged from the initial read. Before/after
SHA256 checks agree for turn.go, identity.go, go.mod, and go.sum:

```text
88f69778c85eebd7139ff292c255741381b28f3b6796ae00966c89eb4a48c401  internal/frame/turn.go
78e20396658bdb15940e62cece042af4334f4f74e85f497dbc0e483ed919689d  internal/frame/identity.go
fa72235136c9ba87b6c33424e3e28c451a53901bd5198db1ee69fe9b4f654da6  go.mod
d6ea4b0e19a07904c297b689eef052f191b7495a1ca8766d6b90594e94243c13  go.sum
```

The benchmark reconstruction is an audit snapshot, not a maintained alternative
extractor. Its source-fidelity test deliberately fails if HEAD or the patch
changes. Benchmark permission does not imply publication approval. Based on
the confirmed lone-delimiter regression and active semantic failures, this
audit does not approve publication.

The final test/benchmark source SHA256 is
`06837fc8afb6e222fbfe414c5c1132a77dc82d7367d33a46500326307538c58d`.
Report readback verified the exact diff and all 360 accepted measurement tuples
against the original command outputs. All audit command sessions completed.

## Original Accepted Measurements

Each JSONL row has six samples in their original reported order. Sample tuple:
`[iterations, ns/op, MB/s, B/op, allocs/op]`. Both timing orders are included,
with no samples dropped. The excluded warm-up run is not mixed into these data.

```jsonl
{"order":"forward","profile":"early","bytesPerRole":256,"version":"HEAD","samples":[[6294,35361,14.48,4374,70],[7224,35198,14.55,4368,70],[6954,34622,14.79,4368,70],[7401,34728,14.74,4368,70],[6974,34951,14.65,4368,70],[7364,36008,14.22,4368,70]]}
{"order":"forward","profile":"early","bytesPerRole":256,"version":"current","samples":[[6342,39793,12.87,9456,94],[6183,39622,12.92,9457,94],[6238,41930,12.21,9457,94],[5547,44101,11.61,9457,94],[6115,49418,10.36,9457,94],[5869,41498,12.34,9457,94]]}
{"order":"forward","profile":"early","bytesPerRole":2048,"version":"HEAD","samples":[[1212,205857,19.9,8678,70],[1203,201164,20.36,8657,70],[1194,202606,20.22,8657,70],[1215,201313,20.35,8657,70],[1219,201025,20.38,8657,70],[1194,206621,19.82,8657,70]]}
{"order":"forward","profile":"early","bytesPerRole":2048,"version":"current","samples":[[999,236276,17.34,49367,94],[1029,246018,16.65,49368,94],[1012,242968,16.86,49368,94],[998,242757,16.87,49368,94],[1009,240733,17.01,49368,94],[1008,242604,16.88,49368,94]]}
{"order":"forward","profile":"early","bytesPerRole":16384,"version":"HEAD","samples":[[132,1879643,17.43,44819,70],[138,1745236,18.78,44764,70],[146,1655906,19.79,44763,70],[148,1624674,20.17,44763,70],[147,1604047,20.43,44759,70],[136,1706737,19.2,44760,70]]}
{"order":"forward","profile":"early","bytesPerRole":16384,"version":"current","samples":[[100,2017988,16.24,368916,94],[100,2069355,15.83,368916,94],[100,2122040,15.44,369620,96],[100,2077422,15.77,368922,94],[100,2141560,15.3,368924,94],[100,2211829,14.81,369371,95]]}
{"order":"forward","profile":"late","bytesPerRole":256,"version":"HEAD","samples":[[4426,55692,9.19,3928,70],[4317,56048,9.14,3928,70],[4330,55540,9.22,3928,70],[4436,55162,9.28,3928,70],[4500,55181,9.28,3928,70],[4488,55213,9.27,3928,70]]}
{"order":"forward","profile":"late","bytesPerRole":256,"version":"current","samples":[[4098,60452,8.47,9112,94],[4134,59951,8.54,9112,94],[4146,59950,8.54,9112,94],[4180,59669,8.58,9112,94],[4140,59813,8.56,9112,94],[4129,59864,8.55,9112,94]]}
{"order":"forward","profile":"late","bytesPerRole":2048,"version":"HEAD","samples":[[555,446360,9.18,8343,80],[559,444414,9.22,8297,80],[560,444337,9.22,8297,80],[550,437934,9.35,8297,80],[556,437389,9.36,8297,80],[541,448147,9.14,8297,80]]}
{"order":"forward","profile":"late","bytesPerRole":2048,"version":"current","samples":[[505,484108,8.46,49007,104],[493,486084,8.43,49007,104],[500,485663,8.43,49007,104],[500,486978,8.41,49007,104],[498,488928,8.38,49007,104],[493,491976,8.33,49007,104]]}
{"order":"forward","profile":"late","bytesPerRole":16384,"version":"HEAD","samples":[[67,3605789,9.09,44508,80],[66,3605040,9.09,44400,80],[68,3600592,9.1,44400,80],[66,3591182,9.12,44400,80],[67,3613040,9.07,44400,80],[64,3600527,9.1,44400,80]]}
{"order":"forward","profile":"late","bytesPerRole":16384,"version":"current","samples":[[60,4337243,7.56,368554,104],[62,4323503,7.58,368561,104],[61,4274921,7.67,368553,104],[60,4347254,7.54,368554,104],[60,4280306,7.66,368555,104],[60,4175028,7.85,368554,104]]}
{"order":"forward","profile":"no-match","bytesPerRole":256,"version":"HEAD","samples":[[1701,140749,3.64,2216,41],[1708,139712,3.66,2216,41],[1682,141507,3.62,2216,41],[1659,140492,3.64,2216,41],[1711,139833,3.66,2216,41],[1713,139553,3.67,2216,41]]}
{"order":"forward","profile":"no-match","bytesPerRole":256,"version":"current","samples":[[1648,147610,3.47,11721,85],[1651,149213,3.43,11721,85],[1647,149901,3.42,11721,85],[1621,149657,3.42,11721,85],[1611,149518,3.42,11721,85],[1656,151640,3.38,11721,85]]}
{"order":"forward","profile":"no-match","bytesPerRole":2048,"version":"HEAD","samples":[[207,1146774,3.57,6621,41],[205,1153850,3.55,6506,41],[210,1137164,3.6,6506,41],[210,1146860,3.57,6506,41],[208,1147028,3.57,6506,41],[211,1140670,3.59,6506,41]]}
{"order":"forward","profile":"no-match","bytesPerRole":2048,"version":"current","samples":[[186,1308774,3.13,81139,85],[184,1304707,3.14,81139,85],[186,1301053,3.15,81139,85],[194,1253155,3.27,81141,85],[188,1259928,3.25,81139,85],[186,1314837,3.12,81139,85]]}
{"order":"forward","profile":"no-match","bytesPerRole":16384,"version":"HEAD","samples":[[22,9575877,3.42,42957,43],[25,9461002,3.46,42622,41],[25,9430888,3.47,42622,41],[25,9496222,3.45,42622,41],[21,10271431,3.19,42627,41],[24,9548260,3.43,42623,41]]}
{"order":"forward","profile":"no-match","bytesPerRole":16384,"version":"current","samples":[[24,10950741,2.99,636894,86],[24,10924536,3,636894,86],[22,10892347,3.01,636905,86],[24,10814799,3.03,636894,86],[24,10871214,3.01,636894,86],[24,11013370,2.98,636894,86]]}
{"order":"forward","profile":"quoted","bytesPerRole":256,"version":"HEAD","samples":[[4338,54288,9.43,2808,69],[4423,54557,9.38,2808,69],[4369,54836,9.34,2808,69],[4503,54550,9.39,2808,69],[4538,54103,9.46,2808,69],[4566,54398,9.41,2808,69]]}
{"order":"forward","profile":"quoted","bytesPerRole":256,"version":"current","samples":[[1848,130903,3.91,11589,96],[1836,131915,3.88,11589,96],[1815,131418,3.9,11589,96],[1869,129036,3.97,11589,96],[1903,133231,3.84,11589,96],[1776,132033,3.88,11589,96]]}
{"order":"forward","profile":"quoted","bytesPerRole":2048,"version":"HEAD","samples":[[548,439267,9.32,7224,79],[552,449550,9.11,7177,79],[542,439468,9.32,7177,79],[549,454055,9.02,7177,79],[548,441021,9.29,7177,79],[552,440863,9.29,7177,79]]}
{"order":"forward","profile":"quoted","bytesPerRole":2048,"version":"current","samples":[[194,1225588,3.34,81183,96],[195,1223450,3.35,81183,96],[195,1224847,3.34,81183,96],[194,1272410,3.22,81183,96],[193,1243829,3.29,81183,96],[192,1252047,3.27,81183,96]]}
{"order":"forward","profile":"quoted","bytesPerRole":16384,"version":"HEAD","samples":[[62,3650378,8.98,43397,79],[66,3619686,9.05,43280,79],[66,3632189,9.02,43280,79],[68,3620614,9.05,43280,79],[63,3653659,8.97,43281,79],[63,3643381,8.99,43281,79]]}
{"order":"forward","profile":"quoted","bytesPerRole":16384,"version":"current","samples":[[22,10843616,3.02,636950,97],[24,10946276,2.99,636939,97],[24,11050582,2.97,636940,97],[24,10833491,3.02,636940,97],[24,12027090,2.72,636939,97],[22,11103295,2.95,636950,97]]}
{"order":"forward","profile":"collective","bytesPerRole":256,"version":"HEAD","samples":[[1796,131429,3.9,2312,42],[1810,130653,3.92,2312,42],[1810,131932,3.88,2312,42],[1842,136003,3.76,2312,42],[1735,150404,3.4,2312,42],[1744,166416,3.08,2312,42]]}
{"order":"forward","profile":"collective","bytesPerRole":256,"version":"current","samples":[[1573,150942,3.39,11801,86],[1693,145262,3.52,11801,86],[1668,146814,3.49,11801,86],[1653,148518,3.45,11801,86],[1646,180393,2.84,11801,86],[1670,146492,3.5,11801,86]]}
{"order":"forward","profile":"collective","bytesPerRole":2048,"version":"HEAD","samples":[[218,1081630,3.79,6711,42],[224,1063007,3.85,6601,42],[226,1052004,3.89,6601,42],[226,1051759,3.89,6601,42],[226,1057183,3.87,6601,42],[226,1053898,3.89,6601,42]]}
{"order":"forward","profile":"collective","bytesPerRole":2048,"version":"current","samples":[[199,1238765,3.31,81221,86],[200,1257495,3.26,81221,86],[206,1232751,3.32,81220,86],[206,1223372,3.35,81220,86],[200,1249368,3.28,81221,86],[194,1220884,3.35,81221,86]]}
{"order":"forward","profile":"collective","bytesPerRole":16384,"version":"HEAD","samples":[[26,8852941,3.7,42979,43],[26,8736173,3.75,42717,42],[26,8787495,3.73,42717,42],[26,8794333,3.73,42717,42],[26,8802276,3.72,42717,42],[26,8822926,3.71,42717,42]]}
{"order":"forward","profile":"collective","bytesPerRole":16384,"version":"current","samples":[[25,10923507,3,636992,87],[25,10215443,3.21,636980,87],[25,10126205,3.24,638452,87],[24,10275262,3.19,636974,87],[24,10324616,3.17,636974,87],[25,10246835,3.2,636969,87]]}
{"order":"reverse","profile":"early","bytesPerRole":256,"version":"HEAD","samples":[[6760,35890,14.27,4368,70],[6795,35993,14.23,4368,70],[6579,36084,14.19,4368,70],[6639,35935,14.25,4368,70],[7050,35934,14.25,4368,70],[6908,36146,14.16,4368,70]]}
{"order":"reverse","profile":"early","bytesPerRole":256,"version":"current","samples":[[5432,40171,12.75,9463,94],[6117,40446,12.66,9457,94],[6295,40227,12.73,9456,94],[6072,40654,12.59,9456,94],[6177,40060,12.78,9457,94],[6178,41664,12.29,9457,94]]}
{"order":"reverse","profile":"early","bytesPerRole":2048,"version":"HEAD","samples":[[1202,207388,19.75,8657,70],[1168,207569,19.73,8657,70],[1166,205488,19.93,8657,70],[1186,205333,19.95,8657,70],[1184,205646,19.92,8657,70],[1177,206421,19.84,8657,70]]}
{"order":"reverse","profile":"early","bytesPerRole":2048,"version":"current","samples":[[984,240557,17.03,49394,94],[1003,241611,16.95,49368,94],[1015,240410,17.04,49368,94],[1028,240718,17.02,49368,94],[1021,239212,17.12,49368,94],[1012,244295,16.77,49368,94]]}
{"order":"reverse","profile":"early","bytesPerRole":16384,"version":"HEAD","samples":[[134,1772558,18.49,44764,70],[138,1747922,18.75,44764,70],[135,1729348,18.95,44760,70],[135,1728686,18.96,45277,70],[138,1703527,19.24,44760,70],[136,1727145,18.97,44760,70]]}
{"order":"reverse","profile":"early","bytesPerRole":16384,"version":"current","samples":[[100,2174499,15.07,368994,95],[100,2190342,14.96,368924,94],[100,2238512,14.64,369371,95],[100,2267180,14.45,368928,94],[100,2252878,14.54,368928,94],[100,2296064,14.27,368928,94]]}
{"order":"reverse","profile":"late","bytesPerRole":256,"version":"HEAD","samples":[[4341,56702,9.03,3928,70],[4286,56962,8.99,3928,70],[4293,57381,8.92,3928,70],[4216,57096,8.97,3928,70],[4302,57474,8.91,3928,70],[4084,58393,8.77,3928,70]]}
{"order":"reverse","profile":"late","bytesPerRole":256,"version":"current","samples":[[3735,61665,8.3,9112,94],[3906,61834,8.28,9112,94],[4038,61743,8.29,9112,94],[4059,61619,8.31,9112,94],[4026,61727,8.29,9112,94],[4088,61690,8.3,9112,94]]}
{"order":"reverse","profile":"late","bytesPerRole":2048,"version":"HEAD","samples":[[537,473046,8.66,8297,80],[531,452151,9.06,8297,80],[550,449513,9.11,8297,80],[535,451814,9.07,8297,80],[540,451398,9.07,8297,80],[504,459839,8.91,8297,80]]}
{"order":"reverse","profile":"late","bytesPerRole":2048,"version":"current","samples":[[462,524816,7.8,49063,104],[464,521967,7.85,49007,104],[468,509544,8.04,49007,104],[468,515535,7.95,49007,104],[474,513642,7.97,49007,104],[470,513902,7.97,49007,104]]}
{"order":"reverse","profile":"late","bytesPerRole":16384,"version":"HEAD","samples":[[63,3762322,8.71,44401,80],[62,3729640,8.79,44401,80],[63,3713862,8.82,44401,80],[62,3727146,8.79,44401,80],[61,3764527,8.7,44401,80],[61,3742108,8.76,44401,80]]}
{"order":"reverse","profile":"late","bytesPerRole":16384,"version":"current","samples":[[58,4380743,7.48,368681,105],[61,4364176,7.51,368562,104],[58,4361565,7.51,368557,104],[61,4368852,7.5,368553,104],[61,4308733,7.61,369703,106],[61,4272883,7.67,369280,105]]}
{"order":"reverse","profile":"no-match","bytesPerRole":256,"version":"HEAD","samples":[[1663,145972,3.51,2216,41],[1659,144818,3.54,2216,41],[1683,143243,3.57,2216,41],[1678,143060,3.58,2216,41],[1665,142920,3.58,2216,41],[1700,142832,3.58,2216,41]]}
{"order":"reverse","profile":"no-match","bytesPerRole":256,"version":"current","samples":[[1560,152768,3.35,11721,85],[1587,152811,3.35,11721,85],[1621,150758,3.4,11721,85],[1610,151941,3.37,11721,85],[1588,151827,3.37,11721,85],[1567,152894,3.35,11744,85]]}
{"order":"reverse","profile":"no-match","bytesPerRole":2048,"version":"HEAD","samples":[[205,1163633,3.52,6506,41],[204,1159186,3.53,6506,41],[206,1160632,3.53,6506,41],[206,1161637,3.53,6506,41],[205,1165205,3.52,6506,41],[207,1153823,3.55,6506,41]]}
{"order":"reverse","profile":"no-match","bytesPerRole":2048,"version":"current","samples":[[186,1291293,3.17,81268,85],[184,1279474,3.2,81139,85],[186,1291917,3.17,81139,85],[184,1299526,3.15,81139,85],[186,1282773,3.19,81139,85],[186,1266088,3.24,81139,85]]}
{"order":"reverse","profile":"no-match","bytesPerRole":16384,"version":"HEAD","samples":[[24,11708017,2.8,44470,43],[22,9971907,3.29,42625,41],[24,9774696,3.35,42623,41],[24,9718913,3.37,42623,41],[24,9889205,3.31,42623,41],[24,9782495,3.35,42623,41]]}
{"order":"reverse","profile":"no-match","bytesPerRole":16384,"version":"current","samples":[[22,11222839,2.92,637237,88],[22,11544017,2.84,636905,86],[22,11172307,2.93,636905,86],[22,10979112,2.98,636905,86],[22,11123792,2.95,636905,86],[21,11340012,2.89,636890,86]]}
{"order":"reverse","profile":"quoted","bytesPerRole":256,"version":"HEAD","samples":[[4389,55640,9.2,2808,69],[4246,55783,9.18,2808,69],[4279,64662,7.92,2808,69],[4372,56087,9.13,2808,69],[4368,55653,9.2,2808,69],[4242,56080,9.13,2808,69]]}
{"order":"reverse","profile":"quoted","bytesPerRole":256,"version":"current","samples":[[1800,133599,3.83,11589,96],[1743,135966,3.77,11589,96],[1722,137560,3.72,11589,96],[1837,137363,3.73,11589,96],[1735,138164,3.71,11589,96],[1758,136297,3.76,11589,96]]}
{"order":"reverse","profile":"quoted","bytesPerRole":2048,"version":"HEAD","samples":[[529,449817,9.11,7180,79],[531,460833,8.89,7177,79],[540,458129,8.94,7177,79],[522,474192,8.64,7177,79],[486,482985,8.48,7177,79],[529,459207,8.92,7177,79]]}
{"order":"reverse","profile":"quoted","bytesPerRole":2048,"version":"current","samples":[[177,1360729,3.01,81319,96],[190,1281923,3.2,81186,96],[183,1303003,3.14,81184,96],[183,1307550,3.13,81184,96],[184,1288773,3.18,81184,96],[189,1282370,3.19,81183,96]]}
{"order":"reverse","profile":"quoted","bytesPerRole":16384,"version":"HEAD","samples":[[60,3849282,8.51,43309,79],[60,3905581,8.39,43281,79],[62,3817798,8.58,43281,79],[61,3856857,8.5,43281,79],[62,3856002,8.5,43281,79],[62,3839508,8.53,43281,79]]}
{"order":"reverse","profile":"quoted","bytesPerRole":16384,"version":"current","samples":[[21,11125762,2.95,637273,99],[21,11320675,2.89,640182,101],[22,11625222,2.82,638965,99],[21,11678129,2.81,636929,97],[21,11474004,2.86,640182,101],[22,11227998,2.92,638964,99]]}
{"order":"reverse","profile":"collective","bytesPerRole":256,"version":"HEAD","samples":[[1758,136005,3.76,2312,42],[1705,137361,3.73,2312,42],[1783,137159,3.73,2312,42],[1724,137794,3.72,2312,42],[1746,136420,3.75,2312,42],[1777,135196,3.79,2312,42]]}
{"order":"reverse","profile":"collective","bytesPerRole":256,"version":"current","samples":[[1648,146418,3.5,11801,86],[1657,146443,3.5,11801,86],[1654,146565,3.49,11801,86],[1646,145029,3.53,11801,86],[1681,144568,3.54,11801,86],[1665,145481,3.52,11801,86]]}
{"order":"reverse","profile":"collective","bytesPerRole":2048,"version":"HEAD","samples":[[217,1098903,3.73,6602,42],[217,1100369,3.72,6602,42],[218,1097325,3.73,6602,42],[214,1108706,3.69,6602,42],[218,1101417,3.72,6602,42],[218,1097615,3.73,6602,42]]}
{"order":"reverse","profile":"collective","bytesPerRole":2048,"version":"current","samples":[[193,1243084,3.3,81343,86],[192,1224292,3.35,81219,86],[196,1234405,3.32,81219,86],[190,1250693,3.27,81219,86],[193,1242866,3.3,81219,86],[193,1241660,3.3,81219,86]]}
{"order":"reverse","profile":"collective","bytesPerRole":16384,"version":"HEAD","samples":[[25,9318092,3.52,42718,42],[26,9396167,3.49,42717,42],[26,9280704,3.53,42717,42],[25,9154147,3.58,42718,42],[25,9309557,3.52,42718,42],[25,9310797,3.52,42718,42]]}
{"order":"reverse","profile":"collective","bytesPerRole":16384,"version":"current","samples":[[24,10790351,3.04,637258,88],[24,10935203,3,636974,87],[24,10654031,3.08,636974,87],[24,10605613,3.09,636974,87],[22,10636479,3.08,636985,87],[24,10512075,3.12,636974,87]]}
```

## Approved Repair And Final Handoff

The two confirmed extraction bugs are repaired. This section supersedes the
historical pre-fix approval status, not its measurements. Semantic validation of
the frozen final source passes. No additional benchmark was started during the
parent's service/final-source runs, and no commands owned by this audit remain
running. Only the four authorized files were edited by this audit; the parent
separately owns its identity_test.go expectation repair. No deployment, commit,
push, branch/index edit, or unrelated cleanup was performed.

### Review Gate And Scope

- Confirmed root cause: regex indices were calculated on a byte-aligned quote
  mask, then the raw capture was promoted without checking whether masking had
  changed any captured byte. Whitespace trimming could conceal the mismatch.
  The full raw span must equal the clean span before trimming or promotion.
- Alternatives reviewed before patching: extracting only clean text would lose
  raw fidelity; filtering later in identity resolution would leave offline
  FromText/query extraction exposed; reordering regex patterns would hide only
  particular examples. The capture-admission boundary is shared by these paths.
- Confirmed actor cause: capitalized pronouns/collective expressions satisfied
  named-actor regexes. Both FromTurn and QueryText now use the same actor
  acceptance predicate. Participant fallback and identity-aware speaker/agent
  rules remain in force; ambiguous collectives cannot become named actors.
- Falsifiers/negative controls: clean named actors and adjacent unquoted clauses
  must still extract; masked-only or crossing captures must not produce direct
  fields. Tests cover both roles, quote forms, multibyte mask alignment, offline
  extraction, query extraction, explicit identities and genuine names.
- This is a reproducible extraction bug, not stored-data corruption. No
  historical-data repair is attempted. Content and normalized What retain raw
  quoted material; the invariant applies to admitted captures, not to archival
  metadata or intentionally inferred fallback text.
- The policy intentionally rejects the entire captured span, even when a clean
  prefix exists. Existing source/pattern priority is preserved; rejection does
  not add a scan for every later match of that pattern. Repeated masks and
  scans remain hot-path costs; no performance refactor is bundled into this fix.

The related-fix check inspected [upstream turn.go](https://raw.githubusercontent.com/JuanHuaXu/eventframed/main/internal/frame/turn.go)
and the [public pull-request listing](https://github.com/JuanHuaXu/eventframed/pulls);
no equivalent repair was found there. The open-issue listing was empty and the
closed-issue query was inaccessible, so that check is not exhaustive.

### Immutable Control And Provenance

The control is the original commit object
`1a7edb62b6be4031fd01ebeab8071b17303a7815`, not necessarily current HEAD.
The audit now reads
`git show turnFallbackAuditHead:internal/frame/turn.go` using that fixed SHA
constant, verifies its exact commit hash with `rev-parse --verify ...^{commit}`,
and verifies its object type with `cat-file -t`. Moving HEAD after an authorized
commit does not invalidate the test. This repository object must remain available.

The complete reconstructed FromTurn, firstField, and participantFallback bodies
remain immutable and are checked by exact normalized AST comparison against that
object. Their original reconstruction was also checked byte-for-byte unchanged.
The AST fidelity checks on all unchanged declarations remain active; only the
three intentionally changed functions and the new actor predicate are excluded
from that unchanged-declaration comparison. Full candidate source hashes cover
the changed declarations. No fidelity assertion was silently removed.

Frozen final SHA256 values:

| File | SHA256 |
| --- | --- |
| internal/frame/turn.go | `88527b6032ebdd1f9d50102134a16fbcddabbf093e0aed708b5e76644fe51cb3` |
| internal/frame/query.go | `34aec61ade88b154686e34efd99096fd4689d04f45d5d1b9de44120b2d4c315f` |
| internal/frame/turn_fallback_audit_test.go | `e9c64e47afb2e7b4a3c732d59cea5405b137d2b882a711c96ca0188967b55a7f` |
| internal/frame/identity.go (untouched) | `78e20396658bdb15940e62cece042af4334f4f74e85f497dbc0e483ed919689d` |
| internal/frame/text.go (untouched) | `06c9d7772ef011a901dccd5f5f2ff7602b64c00f754d473fb38174435c25b66f` |
| go.mod (untouched) | `fa72235136c9ba87b6c33424e3e28c451a53901bd5198db1ee69fe9b4f654da6` |
| go.sum (untouched) | `d6ea4b0e19a07904c297b689eef052f191b7495a1ca8766d6b90594e94243c13` |

Future intentional candidate changes require explicitly refreshing the candidate
hash pins. The immutable control constant must not move. The earlier parent
failure read final turn.go with an older compiled test's intermediate `fe53...`
pin; refreshed final pins and fixed-object fidelity checks pass.

### Final Combined Production Diff

This records the frozen candidate's complete turn.go/query.go diff from the
original control, including the pre-existing user patch plus the authorized
repairs. Existing production files other than these two were not edited.

```diff
diff --git a/internal/frame/query.go b/internal/frame/query.go
index 5bc0605..5713995 100644
--- a/internal/frame/query.go
+++ b/internal/frame/query.go
@@ -17,7 +17,7 @@ func QueryText(query string) string {
 	sources := []sourceText{source}
 	request := firstStatement(source)
 	event := model.Event{
-		Who:  firstField(sources, whoPatterns, .88, nil, model.Field{}),
+		Who:  firstField(sources, whoPatterns, .88, validNamedActor, model.Field{}),
 		What: field("request: "+request.value, model.SourceObserved, 1, request.evidence),
 		Where: firstField(sources, wherePatterns, .86, validLocation,
 			model.Field{}),
diff --git a/internal/frame/turn.go b/internal/frame/turn.go
index d73e55b..2a0e018 100644
--- a/internal/frame/turn.go
+++ b/internal/frame/turn.go
@@ -78,7 +78,7 @@ func FromTurn(turn model.TurnCapture) model.Event {
 		ID: turn.ID, TenantID: turn.TenantID, SessionID: turn.SessionID, Sequence: turn.Sequence,
 		Kind: "agent_turn", Content: "User: " + turn.UserText + "\n\nAssistant: " + turn.AssistantText,
 		OccurredAt: turn.OccurredAt, ObservedAt: turn.ObservedAt, AvailableAt: turn.AvailableAt,
-		Who:      firstField(sources, whoPatterns, .88, nil, participantFallback(turn, user)),
+		Who:      firstField(sources, whoPatterns, .88, validNamedActor, participantFallback(turn, user)),
 		What:     field("request: "+request.value+"; outcome: "+outcome.value, model.SourceInferred, .82, request.evidence+"; "+outcome.evidence),
 		Where:    firstField(sources, wherePatterns, .86, validLocation, field("session:"+turn.SessionID, model.SourceObserved, 1, "session metadata")),
 		When:     firstField(sources, whenPatterns, .90, nil, field(turn.OccurredAt.Format("2006-01-02T15:04:05.999999999Z07:00"), model.SourceObserved, 1, "turn timestamp")),
@@ -92,13 +92,18 @@ func FromTurn(turn model.TurnCapture) model.Event {
 
 func firstField(sources []sourceText, patterns []pattern, confidence float64, accept func(string) bool, fallback model.Field) model.Field {
 	for _, source := range sources {
+		clean := unquoted(source.text)
 		for _, candidate := range patterns {
-			indices := candidate.expression.FindStringSubmatchIndex(source.text)
+			indices := candidate.expression.FindStringSubmatchIndex(clean)
 			group := candidate.group * 2
 			if len(indices) <= group+1 || indices[group] < 0 {
 				continue
 			}
 			start, end := indices[group], indices[group+1]
+			// Captures must not reintroduce bytes hidden by the quote mask.
+			if source.text[start:end] != clean[start:end] {
+				continue
+			}
 			value := strings.TrimSpace(source.text[start:end])
 			if value == "" || (accept != nil && !accept(value)) {
 				continue
@@ -112,15 +117,36 @@ func firstField(sources []sourceText, patterns []pattern, confidence float64, ac
 	return fallback
 }
 
+func validNamedActor(value string) bool {
+	for _, word := range strings.Fields(value) {
+		switch strings.ToLower(word) {
+		case "i", "me", "my", "mine", "myself", "we", "us", "our", "ours", "ourselves",
+			"i'm", "i've", "i'll", "i'd", "we're", "we've", "we'll", "we'd",
+			"you", "your", "yours", "yourself", "yourselves",
+			"you're", "you've", "you'll", "you'd",
+			"he", "him", "his", "himself", "she", "her", "hers", "herself",
+			"he's", "he'll", "he'd", "she's", "she'll", "she'd",
+			"it", "its", "itself", "they", "them", "their", "theirs", "themselves",
+			"it's", "it'll", "it'd", "they're", "they've", "they'll", "they'd",
+			"who", "whom", "whose", "this", "that", "these", "those",
+			"anyone", "anybody", "anything", "everyone", "everybody", "everything",
+			"someone", "somebody", "something", "nobody", "nothing", "none":
+			return false
+		}
+	}
+	return true
+}
+
 func participantFallback(turn model.TurnCapture, user sourceText) model.Field {
-	if indices := collectivePattern.FindStringIndex(user.text); indices != nil {
-		return field("user and agent", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
+	clean := unquoted(user.text)
+	if indices := collectivePattern.FindStringIndex(clean); indices != nil {
+		return field("unresolved group", model.SourceInferred, 0, span(user.name, indices[0], indices[1]))
 	}
-	if indices := firstPersonPattern.FindStringIndex(user.text); indices != nil {
+	if indices := firstPersonPattern.FindStringIndex(clean); indices != nil {
 		return field("user", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
 	}
 	if turn.AgentID != "" {
-		if indices := secondPersonPattern.FindStringIndex(user.text); indices != nil {
+		if indices := secondPersonPattern.FindStringIndex(clean); indices != nil {
 			return field("agent:"+turn.AgentID, model.SourceInferred, .70, span(user.name, indices[0], indices[1]))
 		}
 		return field("user and agent:"+turn.AgentID, model.SourceInferred, .55, "turn participants")
```

### Functional Validation

The earlier sole full-frame failure was
`TestIdentityAuditQuotedMultilineField` at identity_test.go:145: it expected
quoted multiline bytes in directly captured How. Full-span rejection correctly
returns inferred How instead. No regression assertion or capture rule was
weakened to restore the quoted How. The parent repaired the stale test to assert
raw Content preserves the newline quote, What preserves its normalized quote,
How is the .62 inferred `through the agent response: Ready.` fallback, Who is
Alex, and unresolved references are absent.

| Command | Frozen-source result |
| --- | --- |
| `go test -mod=readonly ./internal/frame -run '^TestTurnFallbackAudit' -count=3` | PASS, 0.508s |
| `go test -mod=readonly ./internal/frame -count=1` | PASS, 0.256s, with parent expectation fix |
| `go test -mod=readonly -race ./internal/frame -count=1` | PASS, 2.769s |
| `go vet -mod=readonly ./internal/frame` | PASS |
| `gofmt -d` on the three owned Go files | No differences |

The parent additionally reported final-source frame-race PASS, 2.702s, and source
hash verification. Those are parent-reported results, not a second local run.
Repository-wide/service validation and the parent's final-source benchmark in
runner 45416 are parent-owned; no unobserved outcome or data is claimed here.
All original known-bug regression assertions remain active and pass, including
the masked-capture and collective-actor checks. Final tests additionally cover
contracted pronouns. No production or test files were changed during handoff.

### Intermediate Repair Benchmark (Separate Snapshot)

Measured turn.go SHA256:
`fe53b83f991492c883eedc1c1ce931cd5cb8133020a6a5dac4fbe58b58121746`.
This snapshot includes full-span rejection and named-actor filtering, but
precedes the final contracted-pronoun extension in `88527...`. That extension
is an algorithmic change, not merely a comment edit. These measurements must
not be described as timings of the final frozen source. The parent owns that
final-source benchmark; its data is not duplicated or invented here.

QueryText was already at the final query.go hash above. At measurement time the
harness still checked moving HEAD equaled the fixed control; that equality and
the exact original function reconstruction passed. The final harness instead
reads the immutable object directly. Benchmark labels `HEAD` mean the original
control object, while `current` below means only the intermediate snapshot.

```sh
env GOMAXPROCS=1 go test -mod=readonly ./internal/frame -run '^$' \
  -bench '^BenchmarkTurnFallbackAuditMatchedAlternating$' \
  -benchmem -benchtime=200ms -count=1 -cpu=1
```

Completed PASS in 22.131s. There are 90 accepted readings: 15 fixtures x two
implementations x three repeats. For each fixture, round 1 is control/candidate,
round 2 candidate/control, round 3 control/candidate. Both run complete FromTurn,
including quote handling, on the same deterministic public synthetic two-role
inputs with exactly 256, 2048 or 16384 bytes per role. No private chats are used.
Original quote helpers, unchanged declarations and the full immutable control
are fidelity-checked; the benchmark does not substitute an isolated fallback
microbenchmark. Test setup/fidelity validation are outside timed loops.

Host: Apple M4, darwin/arm64, Go 1.27.1, 10 cores, 16 GiB, GOMAXPROCS=1.
Runs are serial; no exclusive-host, affinity or unrelated-workload guarantee is
claimed. Pre-fix measurements are separately preserved above, not pooled with
these readings or used as a same-time comparison.

Medians of three repeats follow. Latency is time per operation; throughput is
its inverse for fixed input size. Throughput change equals
`100 * (control_latency / candidate_latency - 1)`, not negated latency percent.
B/op and allocs/op are medians; microseconds are absolute measured latency.

| Profile | Bytes/role | Control us/op | Intermediate us/op | Delta us/op | Latency change | Throughput change | B/op control -> intermediate | Allocs/op control -> intermediate |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| early | 256 | 35.966 | 40.266 | +4.300 | +12.0% | -10.7% | 4368 -> 9481 | 70 -> 96 |
| early | 2048 | 203.004 | 243.199 | +40.195 | +19.8% | -16.5% | 8657 -> 49392 | 70 -> 96 |
| early | 16384 | 1730.971 | 2259.119 | +528.148 | +30.5% | -23.4% | 44764 -> 368952 | 70 -> 96 |
| late | 256 | 56.677 | 61.087 | +4.410 | +7.8% | -7.2% | 3928 -> 9136 | 70 -> 96 |
| late | 2048 | 447.293 | 520.247 | +72.954 | +16.3% | -14.0% | 8297 -> 49032 | 80 -> 106 |
| late | 16384 | 3767.657 | 4712.458 | +944.801 | +25.1% | -20.0% | 44409 -> 368588 | 80 -> 106 |
| no-match | 256 | 142.247 | 152.007 | +9.760 | +6.9% | -6.4% | 2216 -> 11721 | 41 -> 85 |
| no-match | 2048 | 1197.220 | 1336.639 | +139.419 | +11.6% | -10.4% | 6506 -> 81140 | 41 -> 85 |
| no-match | 16384 | 9698.277 | 10972.532 | +1274.255 | +13.1% | -11.6% | 42623 -> 636906 | 41 -> 86 |
| quoted | 256 | 55.926 | 134.366 | +78.440 | +140.3% | -58.4% | 2808 -> 11589 | 69 -> 96 |
| quoted | 2048 | 445.871 | 1318.520 | +872.649 | +195.7% | -66.2% | 7177 -> 81184 | 79 -> 96 |
| quoted | 16384 | 3762.952 | 11221.943 | +7458.991 | +198.2% | -66.5% | 43396 -> 636950 | 79 -> 97 |
| collective | 256 | 133.424 | 143.688 | +10.264 | +7.7% | -7.1% | 2312 -> 11801 | 42 -> 86 |
| collective | 2048 | 1077.090 | 1269.284 | +192.194 | +17.8% | -15.1% | 6601 -> 81221 | 42 -> 86 |
| collective | 16384 | 8993.795 | 10326.438 | +1332.643 | +14.8% | -12.9% | 42718 -> 636986 | 42 -> 87 |

Unquoted profiles show median latency increases of 6.9%-30.5% in this sample,
with absolute increases of 4.300-1332.643 us/op. Quoted inputs change semantic
output: formerly successful quoted captures are rejected and more patterns may
be scanned. Their 140.3%-198.2% increases are not a pure equivalent-output
overhead estimate. Comparing pre-fix and repaired ratios across different runs
does not establish causal confidence in the incremental repair cost.

Three repeats are descriptive, not a confidence interval or production forecast.
Noise and all readings are retained. Early/256 round 3 is slightly faster for
the candidate, so the direction is not uniformly stable across every repeat.
No p-value, isolation claim or precise causal percentage is inferred.

| Profile | Bytes/role | Control range us/op | Intermediate range us/op | Paired latency ratios, rounds 1/2/3 |
| --- | ---: | ---: | ---: | --- |
| early | 256 | 34.942-41.145 | 39.865-40.443 | 1.1409, 1.1245, 0.9786 |
| early | 2048 | 201.548-221.615 | 240.428-258.343 | 1.2067, 1.1844, 1.1657 |
| early | 16384 | 1628.862-1771.160 | 2238.158-2263.777 | 1.2781, 1.2930, 1.3869 |
| late | 256 | 56.352-56.731 | 61.014-61.621 | 1.0768, 1.0872, 1.0827 |
| late | 2048 | 442.720-448.900 | 516.506-522.362 | 1.1751, 1.1636, 1.1547 |
| late | 16384 | 3699.200-3818.872 | 4616.424-4753.228 | 1.2616, 1.2739, 1.2088 |
| no-match | 256 | 141.816-142.579 | 150.162-152.032 | 1.0661, 1.0688, 1.0589 |
| no-match | 2048 | 1157.592-1235.720 | 1324.952-1338.128 | 1.1560, 1.0817, 1.1067 |
| no-match | 16384 | 9679.010-9771.250 | 10958.591-11018.778 | 1.1362, 1.1336, 1.1215 |
| quoted | 256 | 55.575-57.069 | 133.790-135.061 | 2.3444, 2.4177, 2.4150 |
| quoted | 2048 | 441.099-455.832 | 1285.323-1350.624 | 2.8926, 3.0620, 2.8827 |
| quoted | 16384 | 3720.940-3764.895 | 11215.552-11387.691 | 3.0159, 2.9805, 3.0247 |
| collective | 256 | 133.153-134.333 | 143.108-143.951 | 1.0726, 1.0791, 1.0716 |
| collective | 2048 | 1074.921-1081.267 | 1254.930-1443.718 | 1.1739, 1.1675, 1.3404 |
| collective | 16384 | 8968.609-9100.240 | 10301.314-10337.951 | 1.1495, 1.1320, 1.1514 |

### Intermediate Raw Measurements

These 90 readings are separate from the preserved 360 pre-fix readings.
Each row's three samples are rounds 1, 2, 3. Tuples retain
`[iterations, ns/op, MB/s, B/op, allocs/op]`; execution order alternates as stated
above. The final frozen-source benchmark is deliberately not represented here.

```jsonl
{"phase":"intermediate-repair","profile":"early","bytesPerRole":256,"version":"HEAD","samples":[[5768,34942,14.65,4374,70],[6805,35966,14.24,4368,70],[6460,41145,12.44,4368,70]]}
{"phase":"intermediate-repair","profile":"early","bytesPerRole":256,"version":"current","samples":[[6226,39865,12.84,9481,96],[6049,40443,12.66,9480,96],[5583,40266,12.72,9481,96]]}
{"phase":"intermediate-repair","profile":"early","bytesPerRole":2048,"version":"HEAD","samples":[[1183,201548,20.32,8679,70],[1178,203004,20.18,8657,70],[1207,221615,18.48,8657,70]]}
{"phase":"intermediate-repair","profile":"early","bytesPerRole":2048,"version":"current","samples":[[934,243199,16.84,49392,96],[1016,240428,17.04,49392,96],[818,258343,15.85,49392,96]]}
{"phase":"intermediate-repair","profile":"early","bytesPerRole":16384,"version":"HEAD","samples":[[136,1771160,18.5,44817,70],[139,1730971,18.93,44764,70],[147,1628862,20.12,44763,70]]}
{"phase":"intermediate-repair","profile":"early","bytesPerRole":16384,"version":"current","samples":[[100,2263777,14.47,368952,96],[100,2238158,14.64,368952,96],[100,2259119,14.5,368951,96]]}
{"phase":"intermediate-repair","profile":"late","bytesPerRole":256,"version":"HEAD","samples":[[4398,56731,9.03,3928,70],[4341,56677,9.03,3928,70],[4198,56352,9.09,3928,70]]}
{"phase":"intermediate-repair","profile":"late","bytesPerRole":256,"version":"current","samples":[[3981,61087,8.38,9136,96],[4000,61621,8.31,9136,96],[4076,61014,8.39,9136,96]]}
{"phase":"intermediate-repair","profile":"late","bytesPerRole":2048,"version":"HEAD","samples":[[544,442720,9.25,8344,80],[548,448900,9.12,8297,80],[549,447293,9.16,8297,80]]}
{"phase":"intermediate-repair","profile":"late","bytesPerRole":2048,"version":"current","samples":[[464,520247,7.87,49032,106],[462,522362,7.84,49032,106],[469,516506,7.93,49032,106]]}
{"phase":"intermediate-repair","profile":"late","bytesPerRole":16384,"version":"HEAD","samples":[[67,3767657,8.7,44517,80],[64,3699200,8.86,44400,80],[67,3818872,8.58,44409,80]]}
{"phase":"intermediate-repair","profile":"late","bytesPerRole":16384,"version":"current","samples":[[61,4753228,6.89,368596,107],[61,4712458,6.95,368587,106],[60,4616424,7.1,368588,106]]}
{"phase":"intermediate-repair","profile":"no-match","bytesPerRole":256,"version":"HEAD","samples":[[1678,142579,3.59,2216,41],[1705,142247,3.6,2216,41],[1683,141816,3.61,2216,41]]}
{"phase":"intermediate-repair","profile":"no-match","bytesPerRole":256,"version":"current","samples":[[1585,152007,3.37,11721,85],[1593,152032,3.37,11721,85],[1599,150162,3.41,11721,85]]}
{"phase":"intermediate-repair","profile":"no-match","bytesPerRole":2048,"version":"HEAD","samples":[[207,1157592,3.54,6621,41],[206,1235720,3.31,6506,41],[193,1197220,3.42,6506,41]]}
{"phase":"intermediate-repair","profile":"no-match","bytesPerRole":2048,"version":"current","samples":[[181,1338128,3.06,81140,85],[180,1336639,3.06,81140,85],[182,1324952,3.09,81140,85]]}
{"phase":"intermediate-repair","profile":"no-match","bytesPerRole":16384,"version":"HEAD","samples":[[25,9698277,3.38,42914,43],[24,9679010,3.39,42623,41],[24,9771250,3.35,42623,41]]}
{"phase":"intermediate-repair","profile":"no-match","bytesPerRole":16384,"version":"current","samples":[[24,11018778,2.97,636906,86],[22,10972532,2.99,638590,86],[22,10958591,2.99,636905,86]]}
{"phase":"intermediate-repair","profile":"quoted","bytesPerRole":256,"version":"HEAD","samples":[[4432,57069,8.97,2808,69],[4488,55575,9.21,2808,69],[4364,55926,9.15,2808,69]]}
{"phase":"intermediate-repair","profile":"quoted","bytesPerRole":256,"version":"current","samples":[[1778,133790,3.83,11589,96],[1801,134366,3.81,11589,96],[1789,135061,3.79,11589,96]]}
{"phase":"intermediate-repair","profile":"quoted","bytesPerRole":2048,"version":"HEAD","samples":[[534,455832,8.99,7225,79],[543,441099,9.29,7177,79],[535,445871,9.19,7177,79]]}
{"phase":"intermediate-repair","profile":"quoted","bytesPerRole":2048,"version":"current","samples":[[183,1318520,3.11,81184,96],[180,1350624,3.03,81184,96],[186,1285323,3.19,81184,96]]}
{"phase":"intermediate-repair","profile":"quoted","bytesPerRole":16384,"version":"HEAD","samples":[[63,3720940,8.81,43396,79],[62,3762952,8.71,43995,79],[63,3764895,8.7,43281,79]]}
{"phase":"intermediate-repair","profile":"quoted","bytesPerRole":16384,"version":"current","samples":[[22,11221943,2.92,636950,97],[24,11215552,2.92,636946,97],[22,11387691,2.88,636950,97]]}
{"phase":"intermediate-repair","profile":"collective","bytesPerRole":256,"version":"HEAD","samples":[[1783,133424,3.84,2312,42],[1783,133153,3.85,2312,42],[1822,134333,3.81,2312,42]]}
{"phase":"intermediate-repair","profile":"collective","bytesPerRole":256,"version":"current","samples":[[1664,143108,3.58,11801,86],[1714,143688,3.56,11801,86],[1620,143951,3.56,11801,86]]}
{"phase":"intermediate-repair","profile":"collective","bytesPerRole":2048,"version":"HEAD","samples":[[222,1081267,3.79,6709,42],[222,1074921,3.81,6601,42],[223,1077090,3.8,6601,42]]}
{"phase":"intermediate-repair","profile":"collective","bytesPerRole":2048,"version":"current","samples":[[198,1269284,3.23,81221,86],[189,1254930,3.26,81221,86],[202,1443718,2.84,81221,86]]}
{"phase":"intermediate-repair","profile":"collective","bytesPerRole":16384,"version":"HEAD","samples":[[26,8993795,3.64,42979,43],[25,9100240,3.6,42718,42],[26,8968609,3.65,42717,42]]}
{"phase":"intermediate-repair","profile":"collective","bytesPerRole":16384,"version":"current","samples":[[24,10337951,3.17,636986,87],[24,10301314,3.18,638519,87],[25,10326438,3.17,636969,87]]}
```
