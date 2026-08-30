# Production Session Replay Results

## Status

This retrospective replay is **inconclusive for confirmation**. The frozen explicit-anchor label produced five design cases across two trajectories and zero confirmation cases. The design measurements are useful diagnostics, not publication-grade evidence.

## Frozen protocol

- Source: read-only OpenClaw `main` agent UUID session logs before `2026-08-28T00:00:00Z`.
- Split: sessions ending before `2026-08-01T00:00:00Z` are design; sessions starting on or after that instant are confirmation.
- Frontier: isolated chronological segments of at most 100 user/assistant text messages; every eligible prior message is nominated.
- Label: a prior message is relevant only when it shares a normalized URL, absolute path, or code-like identifier with the current user message.
- Runtime: 384-dimensional deterministic feature hashing with baseline, update-all Bayesian, and update-all Bayesian plus residual variants.
- Privacy: raw text, extracted anchors, source paths, file names, and production timestamps are not exported. IDs are SHA-256-derived and exported times are trajectory-relative ordinals.

The source manifest covers 137 scanned UUID files, 133 accepted files, and 1,877 accepted messages. Of the accepted files, 104 are in the design block and 29 are in the confirmation block.

## Measurements

| Question | Design result | Result |
|---|---:|---|
| Did the bounded update-all variants cover the eligible frontier? | Nomination recall `1.0`; activation rate `1.0` | Validated for this replay by construction |
| Did update-all improve probability error over the baseline? | Brier `0.44304` to `0.42085`; absolute gain `0.02219` | Exploratory support only |
| Did update-all improve priority-weighted probability error? | `0.44192` to `0.41842`; absolute gain `0.02350` | Exploratory support only |
| Did update-all improve top-10 recall or first relevant rank? | Recall@10 remained `0.90`; MRR remained `0.82` | Not demonstrated |
| Did residual correction improve on Bayesian update-all? | Metrics were identical | Not observed |
| Did the untouched confirmation block validate the gains? | Zero eligible confirmation cases | Inconclusive |

The generated cluster bootstrap interval is not treated as confirmatory because it resamples only two design trajectories. Five labeled cases are too few to support a general performance claim.

## Interpretation

Within this tiny explicit-continuity sample, updating every event in the bounded frontier made the probabilities modestly less wrong, but did not change which memories reached the top ranks. Residual correction did not activate often enough, or did not have enough repeated support, to separate from ordinary Bayesian update-all.

The main finding is about evidence availability: production transcripts do not contain direct usefulness judgments, and exact repeated anchors are uncommon in the post-August holdout. A future prospective collection should record privacy-preserving recall IDs, packed IDs, eventual usefulness feedback, and frozen inclusion probabilities at runtime. That would test semantic memory usefulness without turning later conversation text into an assumed label.
