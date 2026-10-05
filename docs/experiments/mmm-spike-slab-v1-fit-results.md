# Matrix-free spike-and-slab fitting: implementation verified

## Model unchanged by the optimization

The matrix-free profile stores X, W, collapsed-intercept statistics and the
Gram diagonal, not the p-by-p Gram matrix. Each sweep maintains X*E[beta].
Its updates, factor family, prior inputs and coordinate order match the dense
reference. Setup, sweep and bound each cost O(n*p), with O(n*p) storage.
The factor input is copied once per sweep, not once per coordinate. There is
no hidden all-coordinate validation inside each coordinate update.

Maximum discrepancy versus the reference in the tested sweeps is 6.17e-15
for factor parameters and 3.36e-12 for the bound. Comparisons cover 1/7/64
labels, 1/9/255 features, three prior pairs and four coordinate/xi cycles.
Profile ownership tests mutate the original evidence after construction and
verify that fitted updates remain unchanged. Invalid priors, masks, xi and
factor states reject. Saturated inclusion probabilities still fail closed.

## Complete fixed-prior fitter

The frozen component contract is `mmm-spike-slab-v1-fit-contract.md`.
The fitter initializes from the supplied prior and retains the complete
factor/profile state after each sweep, xi update and conditional-intercept
reprofiling. It records all three bound stages and state motion. Stopping
requires both bound and state convergence, not a bound plateau alone.

The 64-label fixtures with 1/9/255 features stop after 6/6/393 cycles;
the 255-feature fixture's final bound is -44.864690798663325 and normalized
state motion 9.99443e-7. These are fitting diagnostics, not predictive scores.
Default test prior pi=.1,c2=1 is a component fixture, not a selected prior for
research-tape evaluation. No hyperprior has yet been adopted.

Tests verify budget-prefix equality, final state/bound agreement, exact label
complement symmetry, independent dense-trajectory agreement, input isolation,
and invalid budgets/priors. All spike-and-slab component and fitting tests
pass under race: 5.209 seconds package time. No fit equations were altered
to obtain a passing test; no failed quality experiment is omitted.

## Performance

Apple M4, darwin/arm64; three benchmark repetitions of three operations each:

| Operation | Time | Allocated bytes/op | Allocations/op |
|---|---:|---:|---:|
| Dense setup+sweep+bound | 11.120-13.794 ms | 2,245,304-2,247,077 | 1,108-1,110 |
| Matrix-free setup+sweep+bound | 0.111-0.155 ms | 147,904 | 14 |
| Complete fixed-prior fixture fit | 74.236-75.536 ms | 58,910,237-58,911,978 | 6,715-6,717 |

Command: `go test ./internal/observationlearners -run '^$' -bench
'^BenchmarkSpikeSlab(ReferenceSweep|FastSweep|Fit)$' -benchtime=3x -count=3
-benchmem -timeout=3m`. Total package time 1.379 seconds.

Allocated bytes are cumulative, not peak RSS. A complete fit includes 393
cycles on this fixture; its cost is not the one-sweep cost. None of these
measurements includes predictive mixture integration, storage, queueing,
concurrent writes, or end-to-end serving. The changed model's fitting time
must not be presented as an equivalent-quality speedup over the Gamma model.

## Next required work

Integrate the actual spike-and-slab predictive mixture, preserving conditional
intercept correlations. Matching only its mean and variance with a Gaussian
would be an additional approximation and must not be silently called the
same predictive law. Freeze a prior or hyperprior contract without selecting
it from consumed outcomes, then build an as-of adapter and quality pilot.
No research-tape quality run, calibration gain or whole-goal completion exists
for this candidate yet. All seven goals remain open.

SHA-256:
- Matrix-free component: `93014a2788e4e5989ccc18199259323475601443c74792821a8b9dd7a8a6b063`.
- Fitter: `9b626817fb300570064484c3f4ada59011b803689818afb1874d5f2659526bbe`.
- Contract: `5dc2802d043e331ca7cf870dca181599df69e01ae3e0c5ae146012bce0ebb457`.

No processes remain running. Production, whitepaper, dependencies, commits
and pushes are untouched.
