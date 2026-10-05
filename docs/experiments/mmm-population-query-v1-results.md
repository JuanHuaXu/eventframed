# Population-input query opportunity

## Result

All 2688 consumed v120 records completed, requiring 20686 posterior fits.
The evaluator integrates over the known generator input distribution and both
possible paid answers. This removes sampled-input and paid-answer luck, but
still knows the teacher and conditions on natural evidence arriving at161.
It is an offline opportunity diagnostic, not an implementable acquisition rule.

Across1344 delayed records, mean population Brier is:

| Arm | Brier |
| --- | ---: |
| No query | .168958621 |
| Random | .167415352 |
| Entropy | .166958523 |
| Original joint value | .167277354 |
| Disjoint joint value | .167300670 |
| Uniform candidate average | .167465994 |
| Population oracle, paid | .164737520 |
| Population oracle, allowing abstention | .164592507 |
| Sample-input oracle, evaluated on population | .165057178 |

Some paid query helps in1064/1344 records; all paid queries harm in221, with59
ties at1e-12 tolerance. Population and sample-input oracle origins differ in331
records (this count includes near-ties). The sample-input oracle loses .000319658
to the population oracle; entropy loses .002221003. Sample-input luck therefore
does not account for the whole oracle gap. No online policy gain is established.

## Evidence and Limits

- [Protocol](mmm-population-query-protocol.md) freezes the distribution and aging.
- [Raw records](mmm-population-query-v1.jsonl) retain both answer branches.
- [Summary](mmm-population-query-v1-summary.json) contains all84 cells and hashes;
  [replay](mmm-population-query-v1-summary-replay.json) is byte-identical.
- [Race contracts](mmm-population-query-final-contracts.txt) cover84 generator/
  schedule fixtures and336 independently enumerated integrals, maximum error
  1.03e-13, plus six branch/isolation fixtures. Distribution pushforward preserves
  mass when multiple raw inputs map to the same feature vector.
- The scorer checks741830 sampled forecast entries against previous separately
  fitted original/counterfactual branches. It does not independently reconstruct
  every full512-input population integral; those rely on the tested Go kernel.
- [Supplemental hashes](mmm-population-query-v1-supplemental-hashes.json) cover the
  source-generation constants omitted from the collection manifest.
- [Run log](mmm-population-query-v1-run.txt):241.68s wall,962.26s user,3.20s system,
  four workers. This is offline collection cost, not serving latency.

Next test the unchanged four critic architectures using population-risk training
targets, retaining actual-answer and sampled-input evaluations and matched-cost
controls. All seven research goals remain open. No production or paper promotion.
