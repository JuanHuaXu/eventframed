# Packet prior v26: served-law component passes; calibration remains open

Date: 2026-10-02. The [frozen protocol](mmm-packet-prior-v26-protocol.md)
PASSES its actual-law and packet-usefulness screens in both independent
splits and both geometries. This closes V25's offline-only wiring gap for
an isolated, fixed-query, single-owner research adapter. It does not complete
any whole goal or authorize deployment. All seven whole goals remain OPEN;
production and the whitepaper are untouched.

## Implementation and Integrity

The [test adapter](../../internal/store/libravdbstore/research_context_prior_v26_test.go)
loads an immutable contextual prior from the original durable pre-feedback
journal. It preserves the underlying ordinary sufficient counts and exposes
`Alpha=2*b_ref+u`, `Beta=2*(1-b_ref)+v` only to a matching caller-derived query
scope. Missing/wrong scope, tenant, policy, epoch, future evidence, unsupported
fractional/working states and unknown keys fail closed. It supports no sharing
or external-writer guarantee. The caller-context contract is research-only,
not a native Store query-context API.

The existing contextual policy uses posterior weight1, identity calibration,
disabled residuals and unscaled rank correction with cap1. No production
service code is changed. The accepted mean reaches the actual BeliefLaw,
PreResidualLaw, CorrectedLaw, ranking, durable journal and returned packet.
All serialized decision fields must agree between served and stored records.
The six diagnostic worlds also reconstruct the same prior from the original
journal and test immutable-origin, wrong-query, no-scope, tenant, unknown-key
and future-label rejection. Fifteen additional validity falsifiers and the
zero-evidence/ordinary-count control pass under `-race`.

The first diagnostic invocation failed to compile because one test omitted
the tenant argument; this was fixed before any cohort. The next invocation
failed a whole-struct comparison of the intentionally nonserialized
`EvidenceGroupKey` (`json:"-"`). Comparing every serialized field confirms
that no forecast mismatch was hidden by that representation difference.
Neither failed pilot is evidence of integration success.

## Artifacts

- [Design](mmm-packet-prior-v26-design.jsonl):256 fresh worlds and8192 outcomes,
  SHA256 `7536ace6380fd8b789c2fa6021ae2d08783834c1dfd240910b3b9504e868d72d`.
- [Confirmation](mmm-packet-prior-v26-confirmation.jsonl):256 different worlds
  and8192 outcomes, SHA256
  `4ca78fe2a32a10d1c893c6ea6c4b89670bf63e417574848b4a71587eb9c21028`.
- [Summary](mmm-packet-prior-v26-summary.json) and
  [independent verifier](../../research/packet-prior-v26-verify.mjs) reconstruct
  76,800 nominee laws, ordinary counts, vector geometry, actual packets,
  expected risks, 19 source/protocol hashes and frozen decisions. Both pass.
- [Race controls](mmm-packet-prior-v26-race.txt),
  [five-package regression and vet](mmm-packet-prior-v26-regression.txt), and
  [caller-scope benchmark](mmm-packet-prior-v26-benchmarks.txt).

Fresh public-number fixtures are synthetic, not agent outcomes. Selection
and omitted-influence certificates explicitly assume coverage. The raw JSONL
contains nominee forecasts and packet records, not full database exports;
durable wire equality is asserted in the source-hashed Go execution.

## Confirmation Screens

Each row is32 independent label worlds. Brier is expected whole-frontier risk,
lower-is-better. Intervals are paired mean +/-3.5SE, descriptive rather than
simultaneous confidence sequences. Recovery requires usefulness gain>=.02
and Brier gain>=.005 with positive lower endpoints; protection permits harm
upper<=.01. The actual served arm is compared to reconstructed baseline
law/order. Bounded rank arms and flat-law arms are offline counterfactuals,
not separate unchanged-service runs.

| Geometry / regime | Baseline -> served usefulness | Usefulness gain or harm | Baseline -> served Brier | Brier gain or harm | Result |
| --- | --- | --- | --- | --- | --- |
| Tight / independent | .475625 -> .689375 | gain .213750 [.148465,.279035] | .407831 -> .378187 | gain .029645 [.025141,.034149] | PASS |
| Tight / reversed | .124161 -> .229547 | gain .105386 [.090548,.120223] | .419258 -> .357744 | gain .061514 [.057749,.065279] | PASS |
| Tight / aligned | .875839 -> .872550 | harm .003289 upper .005364 | .396077 -> .397827 | harm .001749 upper .002046 | PASS within slack |
| Tight / calibrated | .924917 -> .924897 | harm .000021 upper .000039 | .092332 -> .093863 | harm .001531 upper .002051 | PASS within slack |
| Wide / independent | .518750 -> .687500 | gain .168750 [.114225,.223275] | .309455 -> .284250 | gain .025205 [.021397,.029012] | PASS |
| Wide / reversed | .124161 -> .223070 | gain .098909 [.084086,.113733] | .412207 -> .356000 | gain .056207 [.052672,.059742] | PASS |
| Wide / aligned | .875839 -> .871191 | harm .004648 upper .007562 | .213309 -> .215812 | harm .002503 upper .002826 | PASS within slack |
| Wide / calibrated | .923676 -> .923367 | harm .000309 upper .000591 | .187242 -> .189179 | harm .001936 upper .002632 | PASS within slack |

Design agrees: independent usefulness gains .204375/.159375 and reversed
.100587/.100000 tight/wide; all protection endpoints stay below .01. Both
design bounded-rank counterfactuals still fail wide recovery. The bounded
context arm passes that gate in confirmation alone, so its overall rescue
remains unconfirmed. Flat-law forecast safety fails every split/geometry.
The full served arm passes every declared split/geometry screen; this is not
a paired improvement estimate against V25's different cohort.

## Packed Confidence Caveat

The [post-hoc packed audit](mmm-packet-prior-v26-packed-audit.json), reproduced
by [this script](../../research/packet-prior-v26-packed-audit.mjs), changes no
frozen gate. Better whole-frontier Brier and packet usefulness do not establish
good confidence on the selected packet. In confirmation reversed cases:

| Geometry | Mean served forecast | True served usefulness | Served packet Brier | Baseline packet / original baseline-law Brier |
| --- | ---: | ---: | ---: | ---: |
| Tight | .937317 | .229547 | .679297 | .749958 |
| Wide | .910749 | .223070 | .651082 | .748011 |

Thus total packet Brier improves over the uncorrected baseline but confidence
is still badly overstated. Merely changing which items are selected under
the same learned law raises packed Brier by .265417/.239238 relative to the
baseline packet evaluated with that learned law. Those are different
comparators, not contradictory results. The audit reports both and their
paired intervals. One outcome per event has not cured a misspecified prior
or supplied evidence about all unobserved candidates.

Aligned total packet-risk harm upper endpoints .007904/.009720 and calibrated
.000660/.001156 are below .01 in this post-hoc confirmation diagnostic, not
a new prospective safety certificate. A successor must predeclare selected-
packet confidence and priority-weighted proper-risk checks on fresh data.

## Cost and Decision

Maximum observed sequential Recall was17.54ms design and13.65ms confirmation;
outcome calls12.66/14.56ms. They exclude caller scope construction and do not
measure offered-load p99, queue freshness, realistic embedding generation or
large-corpus performance. The separately measured short-query256D scope
construction costs17.199-17.271us, about4.5KB and20 allocations per call.
The V25 pure32-prediction arithmetic benchmark is not a whole-serving bound.

Keep this adapter isolated. Next test selected-packet calibration and a
bounded challenger that can question its baseline anchor using independent,
budget-matched observations. Retain matched-model and misspecified controls;
do not tune on these consumed cohorts. Query changes, visible mutation,
sharing, residual recalibration, loaded freshness, delayed/correlated outcomes,
nonlinear generalization and untouched agent tasks remain open.
