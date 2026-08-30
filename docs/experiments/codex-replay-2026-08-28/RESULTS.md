# Codex Downstream-Use Replay Results

## Status

The frozen retrospective confirmation **supports the bounded update-all Bayesian ranking mechanism for the declared downstream-use proxy**. It does not establish causal improvement in task completion, and the confirmation uncertainty is limited by only three independent source sessions.

## Evidence

- Source: 81 read-only Codex JSONL files before the frozen data boundary; 65 contained completed usable turns.
- Eligible sessions: 13 with more than ten completed turns, split chronologically into ten design and three confirmation sessions before outcome scoring.
- Cases: 1,286 design cases and 138 confirmation cases.
- Independent clusters with at least one labeled case: nine design sessions and three confirmation sessions.
- Frontier size: 10-99 prior completed turns; median 51 in design and 41 in confirmation.
- Label: a prior completed turn is relevant when its explicit anchors match anchors used in a completed downstream tool call or successful patch operation in the current turn.
- Negative control: the same number of positive updates assigned to deterministically shuffled event IDs.

Raw transcripts, tool inputs, outputs, paths, symbols, file names, and production timestamps are not exported. Artifact IDs are SHA-256-derived and times are trajectory-relative ordinals.

## Confirmation results

| Question | Measurement | Result |
|---|---:|---|
| Does EventFrame reduce proper probability error versus baseline? | Brier `0.26144` to `0.24146`; gain `0.01998`, session-bootstrap 95% interval `[0.01042, 0.04155]` | Supported for the proxy |
| Does EventFrame improve top-10 retrieval? | Recall@10 `0.34374` to `0.44346`; gain `0.09971`, interval `[0.04697, 0.33736]` | Supported for the proxy |
| Does EventFrame improve broader retrieval? | Recall@50 `0.87253` to `0.94018` | Supported descriptively |
| Does EventFrame rank the first useful turn earlier? | MRR `0.70707` to `0.87877` | Supported descriptively |
| Is the gain more than generic update/prevalence adaptation? | EventFrame beats shuffled feedback by Brier `0.01606`, interval `[0.00908, 0.02131]`; Recall@10 gain `0.10504`, interval `[0.05270, 0.35641]` | Negative control passed |
| Does residual correction materially improve update-all? | Brier advantage `0.000095`; ranking metrics identical | Not validated |
| Does every calibration measure improve? | ECE worsened from `0.13436` to `0.13783` while Brier improved | Mixed; ECE improvement is falsified here |
| Does the high-priority miss rate improve? | Both baseline and EventFrame were `0.01587` | Not demonstrated |

The shuffled control slightly improved Brier over baseline (`0.00392`) but reduced Recall@10 by `0.00533`. Real outcome-aligned feedback therefore accounts for substantially more than generic posterior adaptation.

## Interpretation

In plain terms, EventFrame learned which earlier turns tended to become useful later. On untouched sessions it retrieved about ten additional percentage points of relevant prior turns in the top ten, and its probability assignments were less wrong under the proper Brier score. Assigning the same updates to random memories did not reproduce those ranking gains.

The residual layer added almost nothing beyond ordinary update-all Bayesian learning in this run. That suggests the posterior update is doing the useful work for this label regime, while residual reuse either lacks repeated action-key support or has too little remaining error to correct.

## Limits

This is a weak-label retrieval experiment, not a causal A/B test. A tool reference shows downstream use, not that recalling a particular turn caused task success. Exact path and identifier matching can also make related turns share broad labels. The next evidence step is a blinded human audit of sampled labels followed by prospective logging of recalled IDs, packed IDs, citations, test outcomes, correction turns, latency, and token use. A randomized near-tie ranking test is still required before claiming causal task-performance improvement.
