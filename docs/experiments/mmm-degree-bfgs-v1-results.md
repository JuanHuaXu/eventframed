# Box-BFGS convergence and incumbent rescue

Numerical convergence improves sharply; predictive quality is not rescued by
this alone. This1008-fit stratified diagnostic covers all phases/cases/schedules,
index0, publications0/128/224, both windows and noise variants. It is consumed
data with repeated initialization fits, not1008 independent trajectories or a
full2688-run quality confirmation. All seven directions remain open.

## Unguarded result

1004/1008 fits (99.60%) reach the unchanged1e-6 projected-gradient tolerance,
versus0/1008 of the matching old16-step fits. All fixed-noise fits converge;
the four capped learned-noise fits repeat the same16-label initialization across
windows/schedules. No Hessian reset or line-search cap occurs in this data screen.

The overall unguarded optimization screen FAILS: three converged learned-noise
fits have worse objectives than the previous method, maximum regression4.103714.
These are different local optima, not proof of global convergence. The frozen
requirement was >=95% convergence AND no objective regressions.

| Variant | Converged | Objective non-harm | Mean Brier gain vs old |
|---|---:|---:|---:|
| Fixed-noise64 | 252/252 | 252/252 | +.00078817 |
| Fixed-noise32 | 252/252 | 252/252 | +.00035794 |
| Learned-noise64 | 250/252 | 250/252 | -.00152769 |
| Learned-noise32 | 250/252 | 251/252 | -.00340817 |

Brier gains average the next32 forecasts per selected publication; positive is
better. These correlated descriptive averages have no new confirmation intervals.
Better Gaussian training likelihood demonstrably does not imply better binary
forecast calibration. Do not interpret solver convergence as forecast accuracy.

## Incumbent rescue

The separately declared post-outcome safeguard keeps the lower training
objective from the old and new fits, choosing BFGS on exact ties. It does not
read future Q/Y or predictive scores. Both attempts are charged. On paired
records it falls back three times, gives1008/1008 objective non-harm by
construction, and retains1001/1008 converged fits (99.31%). Thus the guarded
OPTIMIZATION diagnostic passes; no whole learning goal or broad quality gate
passes by implication.

The Go research wrapper implements both fits, primitive finite-objective
selection, matched model return, and summed work accounting. Its unit tests
match the selected model across all512 inputs. The1008-record guard analysis
uses stored paired fits and the same primitive rule; it is not a production
OpenClaw integration test or a separate full rerun of the guarded Go wrapper.

Guarded mean expected-Brier gains: +.00078817/+.00035794 for fixed64/32,
-.00152133/-.00345537 for learned64/32. Mean total evaluations, including the
incumbent:97.07/95.37 and115.50/111.50. The safeguard repairs local-objective
regressions; it cannot certify future usefulness or prevent calibration loss.

## Verification

- Initial1008-fit collection6.99s, replay6.93s, byte-identical records. These
  test-body times include streamed input parsing, exclude compilation, and are
  not serving latency. Raw traces are preserved, including the unguarded failure.
- Independent pivoted Gaussian elimination verifies all1008 selected fits,
  64512 old/new forecasts,3024 objectives and4536 finite-difference derivatives.
  Maximum errors: forecast1.155e-14, objective1.066e-12, projected gradient
  1.135e-8. It independently reconstructs evidence origins and checks recorded
  future targets against the source. Claimed converged points also satisfy the
  independently estimated tolerance with a2e-7 numerical envelope.
- Known interior and boundary quadratic optima, fixed coordinates,32 principal
  systems, exact one-step Powell damping, indefinite rejection, stress resets,
  nonfinite objective/gradient rejection and incumbent selection pass under race
  detection. Accepted traces are monotone. These are component checks, not a
  global-optimum theorem for the nonconvex likelihood.
- A compile-only fixture shift needed an explicit integer type. A stress test
  initially demanded no resets; repeated negative curvature correctly exercised
  the declared numerical fallback, so the test now verifies identity reset and
  positive definiteness. Separately, the generic optimizer boundary now rejects
  nonfinite callback values instead of potentially declaring false convergence.
  Valid data forecasts remained byte-identical after these local corrections.

## Cost and next action

Three Apple M4 repetitions,64-label fixture: standalone BFGS7.621-7.679ms
fixed-noise,13.558-13.579ms learned-noise. Guarded total11.654-11.842ms and
18.078-18.151ms. Standalone allocations44,016 bytes/3032 allocations and about
267,760 bytes/18785 allocations; guarded totals about50,912/3048 and274,656/18801.
Repeated free-index slice construction in small QP enumeration is a plausible
allocation source, not profiled root-cause proof. No concurrent experiment ran
during benchmarks. Initialization/stack storage excluded. Prediction uses the
unchanged256-feature map and was not separately rebenchmarked.

The numerical safeguard is useful, but additional optimization costs several
times the old fit for modest fixed-noise predictive changes and learned-noise
regression. Do not spend the next package merely polishing this solver or
claiming its sub100ms component time resolves production latency. The next
learning question is composition: can the genuinely new fixed-noise soft
predictors complement the existing Boolean/generic experts through the frozen
arrival-aware comparison machinery? Keep all prior broad controls and account
for fitting cost. The existing full projected-GP tapes may screen composition
cheaply, but converged/guarded deployment claims would require their own full
forecasts and tests. This remains an untested lead, not a promised rescue.

## Evidence hashes

- Raw: `18406784afd62bf315d2be9e7297bc3c5b4ba8e75678fe32a4dc932de73684e7`
- Summary: `cad0f83718dc593984642eb809f5863683da17a7e34fd563088bdfa98dfaee6e`
- Guard diagnostic: `85d17c9bcda94a92ffe599d5e637c6eb51ec7e0f697440a5bd18bae91fff390f`
- Optimizer: `a4bb933960e01874ab8986842d2aa998e5562ada97cd173475e035b9c74af17e`
- Collector: `cfa7a2904b3946e328fd79aae90c6e5811e1e9f01b99ff042d52276cd9b8a709`
- Go guard: `7b1efbe94a4df12d62280f7f1e2007d459868766dbd4262198388e3d755e0f29`
- Summary code: `83614e90c49c90de4a24a0bde9bc3c3e37b53cc47602a1171290ef655da0eb92`
- Independent audit: `d908c6a7ed5db41e4071cbb43fbdb14a2116b8c4b8a39c62a2986a1f52c7d2cb`
- Guard analysis: `b52bc2ca566cf2877b2683dceff69258de371e7bc9e0da50c2e02d9cf1605789`
- Protocol: `846f49d12d1e2a744105c0001aea39b2892538c8ce550005062350c44d1d33ff`
- Guard protocol: `bc2a81536f8cd8e93333d0b12c5325028f60316bbe8ebd603bfc236d19bef87a`

No production, dependency, paper, commit or push changes.
