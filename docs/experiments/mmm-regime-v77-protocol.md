# V77 changing shared-regime mathematical preflight

Freeze before tests. This is a new model lead for recovery/learning, not another
equivalent-computation claim, a completed adaptive-window algorithm or an
empirical rescue. All seven WHOLE success criteria remain unchanged and OPEN.
Previous goal turn PROGRESS: V75 cohort readback, V76 matched timing, source
audit and chained checkpoint. Parent manifest SHA256:
`acd27d5b83186333790ae27fa3c07a27e540b039e6c1e41313ea3f8147b33554`.

## Patch Reasoning Gate

Confirmed symptoms: V74/V75 mean learning remains slower than Adaptive and
has stationary harms; timing changes did not rescue them. Static explanation
weights, insufficient observation, likelihood mismatch and inappropriate
borrowing remain competing causes. No unique cause is established. Older fixed
windows, expert switching and rate-only changes are not this shared-regime model.
No upstream fix exists for this new isolated research prototype.

The new experiment changes the shared explanation itself. A global reset draws
a fresh explanation and all member rates from its conditional prior. A no-reset
branch keeps the explanation and only advances the issuing member's rate. A
mixture of reset histories generally correlates member rates conditional on the
current explanation. Treating those rates as independent is an approximation,
not an exact scalability result. An exact small-state oracle and run-length
mixture test this upstream inference issue before any quality-screen claims.

## One Joint Model

The state is `(H,R_1,...,R_M)`. H has nine classes: three declared mean maps
(baseline, calibration, inverted baseline) times noise rates 0/.1/.2. Their
prior masses are (.1,.8,.1) times (.8,.1,.1). Rate atoms are .02 and .98;
conditional high-atom prior is `(mu_H,i-.02)/.96`, with means clipped to
[.02,.98]. The atom choice is a new bounded preflight model, not V74's full
27-mean/three-family model or a claim of its equivalence.

At each ORIGINAL issue, with probability kappa reset H and every R_i to that
joint prior. Otherwise retain H and all nonissuing rates, and reset the issuing
rate to its H-conditional prior with probability lambda. Then draw ONE outcome
Y from that member's rate. W1 and W2 are independently flipped measurements
of this SAME Y conditional on Y,H. The likelihood sums over Y; it is not the
product of two independent outcomes. Delayed observations replace a factor at
the original issue; arrival creates no transition. Unknown rows still advance
the issue clock without supplying labels.

Three inference methods over that declared model:

- Dense joint-state filter: at most four members, exponential state count.
- Unpruned reset/run-length mixture: conditional member filters per current
  explanation and most recent global reset; exact but component count grows.
- Top-weight capped run-length mixture: explicitly approximate. Record removed
  posterior mass, compare actual joint TV/forecasts with the dense oracle, and
  propagate a conservative error envelope through observations. No discarded
  mass or plug-in small error is an Anti-Pigeon certificate.

A direct conditional-product projection of the dense filter is a separate
negative/approximation control. Preserve measured dependence loss rather than
claiming independence from common explanation membership.

## Tests And Gates

Use kappa 0, 1/16, 1/2 and 1; lambda 0, 1/4 and 1; 1-4 members; independent
prior and transition-sum checks; independent latent-path enumeration on short
histories; delayed first/pair reveals, unknown suffixes and reverse arrivals.
Exact run-length versus dense posterior/forecast tolerance 2e-11. Cap control
must preserve normalization/support and report its error, not pass as exact.
Single member, kappa=0 and kappa=1 are conditional-product limiting controls.

Audit future labels, query/no-publication, same-Y impossible zero-noise pairs,
invalid configurations/duplicates/components/clocks and failed-update atomicity.
No future truth may enter filter transitions, selection or hyperparameters.
Benchmarks run after unit/race/vet, serially. Measure constructor allocation at
150/200 members and bounded cap; distinguish it from RSS, training work and
loaded serving. No complete-core <=400ms claim from a small filter benchmark.

No population confirmation, reserved 2026105409/2026105411 seeds, sealed labels,
private transcripts, production, existing dirty tracked files, whitepaper,
publication, installs or deployments. Freeze source/compiler closure and logs.
Preserve every failed approximation/test. This preflight supplies no useful
split authority, real agent outcomes, non-harm guarantee, durable freshness or
equal-total-cost acquisition superiority.

## Research Sources

[Li, Boyd, Smyth and Mandt (NeurIPS 2021)](https://proceedings.neurips.cc/paper/2021/file/362387494f6be6613daea643a7706a42-Paper.pdf)
motivate latent change hypotheses and bounded inference; this discrete reset
model is our independent construction, not their Gaussian tempering algorithm
or an inherited guarantee. [Bifet and Gavalda](https://www.cs.upc.edu/~Gavalda/papers/adwin06.pdf)
motivate evidence retention that changes with drift; their independent bounded
stream assumptions do not automatically hold for selected delayed evidence.
Do not label this model ADWIN, ordinary full-history exact Bayes when capped,
or a validated EventFrame adaptation rescue.
