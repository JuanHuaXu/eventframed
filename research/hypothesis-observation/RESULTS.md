# Explicit hypothesis observation v1 results

2026-09-12. [Protocol](PROTOCOL.md). Machine-readable evidence is in
`docs/experiments/mmm-hypothesis-v1-summary.json` and the matching `.json.gz`.
2,560 episodes (256 per case per split), four paired policies,16 queries each.

## Verdict

Matched-model mechanism PASSED. This is not general-use success: a finite known
hypothesis family and known likelihood table are supplied, not discovered, and
the correlated-noise stress case violates the independent-repeats likelihood.
No production or real-world claim is established.

Confirmation mean pre-query target-class Brier across16 steps (lower is better):

| Case | Random | Label entropy | Full-H information | Target Gini |
| --- | ---: | ---: | ---: | ---: |
| Noise05 | 0.296058 | 0.245460 | 0.214943 | 0.115554 |
| Noise20 | 0.557019 | 0.558448 | 0.352116 | 0.284453 |
| All relevant | 0.407318 | 0.358325 | 0.229872 | 0.233809 |
| Uninformative | 0.750000 | 0.750000 | 0.750000 | 0.750000 |
| Correlated noise | 0.363375 | 0.221774 | 0.272963 | 0.215012 |

Primary gains against random are0.180504 [0.127815,0.233193] and
0.272566 [0.198993,0.346138]. Against label entropy they are0.129906
[0.052044,0.207769] and0.273995 [0.206452,0.341537]. Intervals use the frozen
approximate paired z=3.3 rule. All exceed the .02 gain requirement. Target Gini
also improves over full-H information in the two primary cases, but not in the
all-relevant case. It is not a universal dominance result.

## Scope and failure modes

Target Gini final accuracy is100% on noise05 and98.05% on noise20. These are
four-class diagnostic scores, NOT MMM's earlier94.7% event prediction measure.
Do not combine their denominators or metrics. In uninformative episodes all
posteriors remain uniform; no policy manufactures certainty.

Correlated-noise final target-Gini accuracy is91.41%, with2/256 confidently
wrong decisions (0.78125%, max class posterior>=.9). Random has16/256 and
label entropy22/256; full-H information2/256. This finite count is not a
calibration certificate and zero counts elsewhere would not establish zero risk.
Treating repeated source evidence as independent remains invalid even if the
aggregate Brier is better than a comparator.

Next lead: explicit provenance-aware repeated-test likelihoods, or suppress
repeated observations from the same source when they add no conditional
information. Unknown dependency must not be silently declared independent.
The observer also needs a real mechanism for constructing/validating hypothesis
families and mapping actual observations into likelihoods before agent use.

## Verification

Unit tests passed for normalization, fair-coin no-update, relevant versus
nuisance first-query controls, and pre-outcome journal reconstruction.
`verify_evidence.py` regenerated all records and verified exact equality,
source/protocol hashes, seed uniqueness, decisions and summaries. Verification
uses the same implementation, not an independent proof of every calculation.

The acquisition API receives only posterior, likelihoods and class grouping;
simulated truth and future outcomes remain outside it. A generated outcome is
consulted only after the test is selected. Every policy gets the same16-query
budget and common potential-outcome tape. No private data was used.

Research inspiration: [Golovin, Krause & Ray](https://arxiv.org/html/1010.3091).
Our expected target-Gini reduction is not EC2 and inherits none of its
approximation guarantees. Known target classes are a strong declared assumption.
