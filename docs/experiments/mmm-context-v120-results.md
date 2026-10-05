# Context-dependent expert routing: no rescue

**FAIL.** The contextual predictor passes protection against its matched global
model but yields zero qualifying recovery gains over it. Extra context routing
is not justified on this workload. These are consumed v120 data, not fresh
confirmation, a production integration or a completed research direction.

## Evidence

- [Frozen protocol](mmm-context-v120-protocol.md)
- [Exact finite model and cost accounting](../../research/context-expert-router-model.md)
- [Component checks](mmm-context-v120-component-checks.json)
- [Every record, group and gate](mmm-context-v120-diagnostic.json)
- [Independent global reference](mmm-context-v120-global-reference.json)

All2688 runs and21 cases are retained. There are1,376,256 forecast evaluations
and672 poisoned-prefix checks. Input bits, issued expert laws and eligible
outcomes are identical across the contextual and global arms. The contextual
model integrates masks and per-cell expert assignments; it does not use a
teacher-derived best-expert label. Neither model uses the initial16 source-fit
examples as extra meta-training examples without issued-law records.

The two-bit/two-expert component is checked against exhaustive latent expert
assignments over all64 outcome and64 availability patterns, giving32768
comparisons. Maximum probability/evidence errors are8.88e-16 and1.78e-15.
Duplicate/conflict handling, expiry, ownership, probability boundaries and
revival after normalized-weight underflow pass. An independent full global
reconstruction covers688128 forecasts: maximum aggregate score error6.11e-16
and final log-evidence error1.28e-13. This is not independent reconstruction
of every nine-bit contextual forecast. The complete diagnostic reproduces
byte-for-byte, with source hashes recorded.

## Outcomes

| Comparison | Protection | Recovery gains |
| --- | ---: | ---: |
| Contextual versus matched global | 168/168 | 0/32 |
| Contextual versus Markov12 | 155/168 | 1/32 |
| Global versus Markov12 | 155/168 | Not a candidate gain claim |

Thus the contextual candidate passes324/400 requirements, not the whole gate.
The single Markov gain is design-phase immediate additive-gradual, not a
replicated broad improvement. Protection uses the permitted .01 upper loss
increase, not zero harm. Gains require mean>=.005 and positive lower endpoint.
Intervals are the frozen paired mean +/-3.5SE over32 trajectories, not
simultaneous or anytime coverage.

Confirmation delayed terminal64 expected Brier and mean non-global mask mass:

| Case | Markov | Global | Contextual | Non-global mass |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221755 | .217550 | .217456 | .069024 |
| Additive gradual | .239065 | .234260 | .234315 | .051219 |
| Parity4 stationary | .049066 | .048304 | .048304 | .000119 |
| Majority to parity | .060262 | .061924 | .062326 | .014199 |
| Parity to majority | .102272 | .108178 | .108185 | .021117 |

In the reverse switch, contextual gain versus global is -.00000628 with
interval [-.00039276,.00038021]. Against Markov it instead increases loss
.005912, interval [-.009696,.021521], failing the upper harm ceiling. A small
mean increase is not a passed uncertainty criterion. The low non-global mass
is descriptive, not proof that a more permissive prior would help.

## Next decision

Do not spend a fresh confirmation seed on this failed unchanged candidate or
tune context priors until consumed cases pass. The [residual-logit lead](../../research/residual-logit-lead.md)
proposes calibration of issued expert forecasts, with a baseline-only control,
as a different combination family. It is neither implemented nor validated.
Its signed corrections can escape the probability hull and can also overfit;
all existing stationary and switching controls remain required.

No serving timing was measured. Context routing costs O(K*2^d) per prediction
and accepted outcome, with O(K*3^d) cell state plus lookup metadata. Correct
finite inference alone does not justify that cost. Go, production, OpenClaw,
the whitepaper and published repositories are unchanged.
