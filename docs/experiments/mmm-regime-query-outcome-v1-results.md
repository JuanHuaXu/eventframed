# Joint-query actual-outcome diagnostic results

Status: **no established advantage over random or entropy selection**. This was
the frozen all-case one-decision diagnostic, not a whole-stream pass/fail gate.
All seven complete research directions remain open. Do not promote the policy.

## Scope and result

All2688 schedule runs were collected:21 cases, two consumed phases,32 indices,
and paired complete/delayed schedules. Decisions were fixed at160, paid evidence
revealed at161, and actual fitted forecasts scored on161..191. All four arms
used the same publication model, natural feedback and64-label cap.

Equal-cell descriptive means over the42 delayed phase/case cells (1344 runs):

| Policy | Actual expected Brier | Model-predicted single-probe-time gain | Paid labels | Redundant purchases |
| --- | ---: | ---: | ---: | ---: |
| No query | .168933 | n/a | 0 | 0 |
| Random | .166980 | .004818 | 1344 | 32 |
| Query-label entropy | .166657 | .005322 | 1344 | 38 |
| Joint predictive value | .167328 | .007101 | 1344 | 31 |

The actual mean Brier gain for joint value is+.001605 versus no query,
-.000348 versus random, and-.000671 versus entropy. These descriptive averages
are not a population guarantee. Of42 delayed cells, three have positive
mean-minus3.5SE gains versus no query. None has a positive lower gain bound
versus random or entropy; none has a negative upper gain bound against either.
That does not establish equality, nor does it establish broad harm.

Phase1 examples (lower Brier is better):

| Case | No query | Random | Entropy | Joint value |
| --- | ---: | ---: | ---: | ---: |
| Stationary case0 | .216769 | .217059 | .216107 | .217400 |
| Majority to parity19 | .225168 | .217887 | .215526 | .215330 |
| Parity to majority20 | .204004 | .198450 | .201056 | .205009 |

All1344 complete-delivery runs have zero paid cost and identical four-arm
forecasts. In delayed runs, paid evidence changes training support in1312 random,
1306 entropy and1313 joint-value runs. Thus the joint arm's purchases reach the
fitter in97.69% of runs; only31/1344 (2.31%) are already naturally available at
reveal. Merely wiring answers into fitting does not establish superior selection.

The model-predicted gain column is not directly calibrated to the actual Brier
column: it uses eight virtual next-time probes, whereas the scored publication
also admits new natural feedback, can change support, and ages across31 frames.
Do not describe the difference as a numerical failure of the Bayes identity.

## Verification and cost

- [Protocol](mmm-regime-query-outcome-protocol.md),
  [raw records](mmm-regime-query-outcome-v1.jsonl),
  [summary](mmm-regime-query-outcome-v1-summary.json).
- [Race contracts](mmm-regime-query-outcome-contracts.txt):12 fixtures cover
  decision/publication poisoning boundaries, complete-delivery identity, paid
  support inclusion, same-unit costs, redundant-purchase accounting, detached
  replay, concurrent replay and ownership. Six paid-label flips change forecasts.
  Package68.001s. The earlier tape view was factored before fitting to avoid an
  unused reference fit; its pre-change source is archived in
  `research/regime-query-pre-outcome-tape.go.txt`.
- The original tape-adapter race test was rerun after factoring:
  [regression log](mmm-regime-query-tape-after-outcome.txt). All12 views and six
  delayed query values pass, preserving its earlier as-of/ownership contract.
- Independent scorer audits all333312 emitted probabilities' scoring/aging,
  all query selections, marginal/conditional identities, supports, costs,
  redundancy and actual fit counts. This is not independent reconstruction of
  all large-history fitted laws. Existing fitter and batch numerical contracts
  remain the numerical evidence for those components.
- Arithmetic controls and eight deliberately corrupted records are accepted/
  rejected as expected. Raw/source/code hashes are checked and recorded.
- Full Go recollection is byte-identical for all2688 records. Independent summary
  replay is also byte-identical. No row or failure was dropped.
- First [collection](mmm-regime-query-outcome-v1-run.txt):90.06s wall/355.00s user,
  four workers,1344 shared query batches and5886 actual publication fits.
  Identical supports reuse a fit within a record; that count is audited.
  [Full replay](mmm-regime-query-outcome-v1-replay-run.txt):89.20s wall/351.55s user.
  These are experiment costs, not deployed serving latency. A shared batch
  supplies the three selectors; no claim of three free independent batches.

Intervals are exploratory fixed-sample mean +/-3.5SE over32 trajectories within
each cell. Both phases are consumed. No empirical rescue or broad claim closes.

## Next mechanism to isolate

A post-hoc [exact-input exposure diagnostic](mmm-regime-query-probe-overlap.json)
uses only source inputs and recorded selections, not query/future Y or Q:

| Selected policy | Query origin itself among probes | Probe mass on exact selected input | Actual future mass on that input |
| --- | ---: | ---: | ---: |
| Random | 87.95% | 11.21% | .211% |
| Entropy | 92.71% | 11.75% | .238% |
| Joint value | 100% | 13.04% | .240% |

The virtual probe window includes seven of the eight candidate positions. The
joint selector always chooses one of those positions here. This is legitimate
under its declared objective, but markedly unlike the future input frequency in
this workload. Exact-input overlap is not the learner's masked covariance and
does not prove the cause of weak learning. A controlled probe change is needed.

Next test the [disjoint-probe control](mmm-regime-query-disjoint-protocol.md)
without changing evidence, candidate pool, cost, priors or publication. Keep
support/natural-arrival changes as competing explanations, not settled causes.
No production, OpenClaw, dependencies, paper, commit or push changes.
