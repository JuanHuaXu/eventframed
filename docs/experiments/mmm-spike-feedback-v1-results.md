# Delayed-feedback mixture: useful but incomplete rescue

Contract: `mmm-spike-feedback-v1-contract.md`. Implementation:
`research/spike-feedback-v1.mjs`. Consumed index-0 cadence predictions only;
84 records, 2,688 forecasts, 1,869 unique arrived feedback updates. No model
was refitted or selected using evaluator probabilities in this experiment.

| Method | Mean expected Brier |
| --- | ---: |
| Markov incumbent | .178978294 |
| Arrival-refitted challenger | .178917923 |
| Fixed half-mixture | .174272111 |
| Delayed-feedback mixture | .172047334 |

The feedback mixture improves 43/84 records against Markov and improves all
four pooled phase/schedule cells in both expected and realized Brier. This
is a descriptive .006930961 average expected-score improvement, not a
confidence-certified effect. Some benefit comes from diversification alone:
the fixed half-mixture already improves the pooled score.

**The incumbent-protection screen fails.** Six records worsen by more than
.01 versus Markov, compared with fifteen for the uncombined challenger:

| Phase | Scenario | Schedule | Mixture minus Markov expected Brier |
| ---: | --- | --- | ---: |
| 0 | mux3 | delayed | +.035111365 |
| 1 | additive stationary | delayed | +.021502968 |
| 0 | additive stationary | delayed | +.015192224 |
| 0 | mux3 | immediate | +.013225159 |
| 1 | local-table stationary | delayed | +.012133059 |
| 1 | local-table stationary | immediate | +.010922747 |

No unconditional promotion is supported. These 32-forecast blocks provide
neither broad generalization nor a long-run stability result. They do show
that combining complementary experts is a viable lead, rather than requiring
each challenger to replace the incumbent outright. All seven goals stay open.

## Verification and cost

Hand-derived updates, exact-once feedback, missing-label exclusion, delayed
arrival, zero-feedback behavior and future/current-label poisoning tests pass.
Independent batch recomputation of each weight agrees to 2.22e-16. Replay
reproduces every forecast, update and summary exactly; elapsed milliseconds
are the only different JSON field. Processing took .932s and .866s including
reading the source and artifacts, excluding output serialization. These are
postprocessing times, not end-to-end latency. The method still incurs all
2,101 challenger fits and 2,688 challenger forecasts, plus incumbent cost.
The diagnostic update loop is O(32^2) per record, with bounded 32-row state;
it is not an optimized serving implementation.

Full existing spike-model race suite rerun: PASS, 43.865s. Production tracked
diff remains the same four preexisting files (+63/-1); this experiment changes
only research scripts, documents and artifacts.

SHA256:

- Script: `8a8c1f4b1315b8ab5833e393c6e9f40125fdba92a7e8d91647fe5f5ee6e8a991`
- Contract: `f377ba4a35e1490ba6fb040c06e328b0dd038ac92cb57200dbac8b4e19717b2d`
- Result: `d048596606c3846f50ce8ef11d72cd69d7b38b1ac7dcba009e13c8457c998548`
- Replay: `4ff09a2608f67f8fe6f1808d400b782543a13df5790df0db6889450f4652a774`

Next candidate, not yet implemented: a conservative budget on cumulative
realized excess loss, accounting for worst-case excess on pending/missing
outcomes before issuing a mixture. For squared loss, excess at probability
p versus b is `(p-b)*(p+b-2*y)`, whose maximum over binary y occurs at an
endpoint. This permits a predictable bound without consulting hidden labels.
Freeze the budget and release policy before testing; show whether protection
retains useful adaptation rather than merely collapsing to the incumbent.
A pathwise realized-loss bound is not a per-realization expected-Brier or
population confidence certificate. Keep those requirements distinct.
