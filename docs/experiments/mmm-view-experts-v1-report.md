# Separately scored view experts: partial benefit, recovery FAIL

## Result

All128 consumed trajectories, both schedules,512 forecasts each. The frozen
[contract](mmm-view-experts-v1-contract.md) keeps component advice, fitting,
acquisition and Anti-Pigeon clocks unchanged. Only a new outermost selector
learns narrow-versus-available view weights from issued forecasts on label arrival.

**0/4 reverse recovery screens;15/16 full/post nonharm screens. Not adopted.**
The second-cohort delayed reverse case fails nonharm. No goal is completed.

Delayed post-window Brier (lower is better):

| Cohort / case | Narrow control | Blind available | View selector |
| --- | ---: | ---: | ---: |
| 1 / stable majority | .048535 | .054847 | .048514 |
| 1 / stable parity | .144882 | .086316 | .084690 |
| 1 / majority to parity | .273706 | .269874 | .273016 |
| 1 / parity to majority | .259748 | .266852 | .261421 |
| 2 / stable majority | .051261 | .058518 | .051450 |
| 2 / stable parity | .183083 | .099076 | .098917 |
| 2 / majority to parity | .270235 | .266394 | .269732 |
| 2 / parity to majority | .252457 | .269253 | .258092 |

Stable parity accuracy in cohort2 rises65.55% to87.35%. Stable majority remains
94.68%. Average available-view weight is90.95% versus4.61%, respectively. Thus
the same frozen selector can distinguish useful views in stable cases without
blindly applying the expanded law everywhere. That component benefit does not
establish adaptive recovery: delayed reverse mean gains are-.001673 and-.005635,
with exploratory mean+-3.5SE intervals[-.005965,.002619] and[-.010824,-.000445].
Intervals use16 trajectories, not independent frames or anytime coverage.

## Interpretation and next action

The new selector operates on two already-mixed laws. It cannot repair poor
constituent advice by assigning different weights inside those laws. The
post-hoc [component diagnostic](mmm-view-experts-v1-components.json) separates
early and late post-shift windows without changing any model:

- Late delayed reverse short-inner Brier is .23855/.22483 on narrow inputs and
  .23475/.22763 on available inputs, below the neutral .25 baseline.
- Early delayed reverse short-inner Brier is .28499/.26809 narrow and
  .27901/.27449 available, above .25. The apparent weakness over the full window
  is partly slow adaptation, not permanent absence of useful advice.
- Late delayed forward short-inner remains above .25 on both views: available
  .26325/.25892. Forward and reverse failures need not have the same cause.
- These are retrospective component means, not an oracle deployment rule,
  a proof that mixtures cannot improve, or an untouched hypothesis test.

Next distinguish model learning delay from inner/outer credit assignment using
issued constituent advice and arrived training samples. Inspect earlier selector
and challenger rescues first. Do not simply tune the view selector's prior/share
on this consumed tape or use the true switch clock to reset it. Expanded short
advice still inherits the narrow-trained inner mixture; that coupling is unresolved.

## Verification

Race contract test passed (1.267s package); vet passed. Contracts cover future
outcome isolation, missing-label isolation, stable simultaneous-arrival order,
same-advice identity and invalid timing/probability rejection. Prediction occurs
before deliveries at the same clock. Late pre-split advice remains eligible by
the predeclared persistent-view-role policy, not accidental epoch omission.

Collection1.45s; complete race collection23.24s. Both complete raw artifacts are
byte-identical. Independent JavaScript reconstruction checks786,432 scalar
forecast, weight and original-law values, maximum difference8.88e-16. Six focused
source/contract snapshots verify against disk. Summary rerun is byte-identical.
This does not claim a snapshot of every repository dependency; parent tapes have
their own preserved source snapshots and are bound by exact hashes.

The retained-result forecast-plus-update microbenchmark on Apple M4 measured
36.44/36.38/36.54ns per step,0B and0allocations. This is four-slot arithmetic,
not serving latency, database cost, acquisition cost or model-fit time. The
replay scheduler is O(n+maximum delay) work/storage with fixed four-slot updates;
production bounded journal design is not implemented by this diagnostic.

## Artifacts

- Raw: `mmm-view-experts-v1.jsonl`, SHA256
  `4fae78431f66b86de27f3c71458de89dca7605e2ad39bb83bab868de0d9afeeb`.
- Full race replay: `mmm-view-experts-v1-race.jsonl`, same SHA256.
- Summary: `mmm-view-experts-v1-summary.json`; independent evaluator
  `research/view-experts-summary.mjs` is captured in the raw header.
- Input: verified available-evidence tape SHA256
  `3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e`.
- Parent: credit-learning tape SHA256
  `4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655`.
- Post-collection benchmark source SHA256
  `b7accf7f8397c0194c7289cb769446bfc6f09473acd4aa5a1cc4016fc3c11d6c`.
- Post-hoc component diagnostic script SHA256
  `1fbb72fa6702fe0257cfa2b16ee12ab55137a572aed27b68f4012c6448bb64ec`;
  component output SHA256
  `0d9a2ddf284c3d9c7d88d5214a95cfe8192004a853d45dcf62faa2a94cb4c979`.

Everything is isolated synthetic research. Production, whitepaper, remotes and
existing tracked modifications are unchanged. No fresh confirmation, calibrated
belief guarantee, equal-cost observation-policy win or end-to-end utility claim.
