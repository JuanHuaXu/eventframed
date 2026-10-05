# V68 Coherent Shared-Context Diagnostic

Prospective diagnostic on CONSUMED worlds, not confirmation. No adoption,
production, private data, sealed labels, paper edits or publication. All seven
whole goals remain open. The previous research turn made progress; intervening
social clarification was not research progress. This study tests parts of
goals 1/2/4/7, not the separate requirements of 3/5/6.

## Model And Research Basis

One declared joint family contains three alternatives: independent local rate
and local noise (L); independent local rates with one shared static noise (N);
one shared contextual calibration field with shared static noise (S). Hybrid
mixes these with prior (1/3,1/3,1/3), updated by their actual visible joint
evidence, not by grading forecasts. Local and noise modes disable the other
alternatives, rather than relabeling the mixed output.

Local states retain V66's 22 atoms, baseline-preserving .8 baseline plus .2
beta-binomial(20,b,1-b) grid prior, with 1/16 reset at each member ordinal.
Noise hypotheses are {0,.1,.2}, prior {.8,.1,.1}, and never reset.
Shared field: 16 atoms, .72 baseline plus three loading families with masses
.12/.08/.08. Their loadings are 1, 2b-1, and 2(2b-1)^2-1; offsets are
{-8,-4,0,4,8}, weights {.01,.09,.8,.09,.01}. Each family's logistic intercept
is solved before data so its prior mean equals b. The field resets with
probability (1/16)/M at each global issue, not at every receipt. Static noise
stays conditional in normalized field vectors, with log evidence kept separate
to prevent numerical extinction after long agreeing streams. Zero likelihood
under noise=0 after disagreement is a legitimate exclusion, not a floor.

The implementation differs from the earlier four-member toy, which used 30
states and narrower offsets. This 48-state field is declared before collection.
The pure shared model does not represent every possible rate vector; the hybrid
includes the independent alternative with full local grid support. No external
truth, safe grouping, causal evidence or Anti-Pigeon certificate is implied.

W1 and W2 measure the SAME Bernoulli Y. Late W2 replaces the first-only emission
at its origin, not at arrival. Global and local issue clocks are distinct. Clean
and W1 issue laws are recorded before issue; delayed receipts use only visible
evidence. Missing W2 supplies no value. Journal caps: 2..200 members, 64 trials
each, checkpoint stride 128; no 600-frame expiry. Serialized owner only.

Predictive acquisition averages future Brier value over ALL affected members:
sum_j sum_w P(W2=w|visible)*(q_j^w-q_j)^2/M. Every target advances by one next
global issue in the shared model, or its next local ordinal in a local model.
This one-step model value is not a guarantee of long-term external utility.
Tower identity tolerance 2e-10. Information/Gini use the probe marginal
(alternative, origin noise, origin state), with 180 padded atoms, not the entire
latent history or an empirical safe-sharing certificate. Uncertainty is W2
entropy. Random does not compute every candidate score.

[Bonilla, Chai and Williams (2007)](https://proceedings.neurips.cc/paper/2007/file/66368270ffd51418ec58bd793f2d9b1b-Paper.pdf)
motivates testing transfer and explicit non-transfer alternatives: inter-task
sharing can help or harm. Our finite Bernoulli model is NOT their Gaussian
process and inherits none of its empirical results or complexity guarantees.
[Adams and MacKay (2007)](https://arxiv.org/html/0710.3742v1) motivates a declared
reset clock; its assumptions do not establish our stream's change law.

## Experiment And Unchanged Gates

All 40 consumed paired worlds, seed base 2026105407, two generators, 20 regimes,
three delays (immediate/fixed150/uniform299), 2,400 issued packets per arm.
14 configurations/policies: Full, Adaptive; hybrid no_pair/random/uncertainty/
information/falsification/predictive; local, noise, shared each no_pair and
uncertainty. Exactly 25 W2 requests per round in paired policies: 400 per arm.
1,680 arms share 96,000 underlying Y outcomes, not independent arm labels.
No reserved seed or untouched outcome labels opened.

Unchanged metrics: clean expected issued/priority Brier, terminal Brier, top-10
usefulness, recovery. Observed W1/W2 Brier is separate. Gates: >=.01 mean risk
gain vs Full; no cell >.01 Adaptive harm; faster mean recovery vs Adaptive;
complete elapsed core <=400ms in every cell; constructor allocation <=8MiB at
BOTH 150 and 200 members (not RSS). Report all failures and all modes, no pruning.
Core elapsed includes setup, scheduling, issue, all queries/sorting, requests,
delayed/missing receipts and snapshots. Audit/reference time is separate.
Equal request count is NOT equal TOTAL cost; goal 7 needs a cost-matched test.
No inference about durable persistence, loaded RPC or sub-100ms serving.

## Audit

Freeze compiler closure/generated test mains, scripts, protocol, parent and 14
protected dirty file hashes before collection. Independent dense transitions,
urn prior reconstruction, enumerated Y emissions, evidence ratios and smoothing
check the candidate without importing its package. Race, vet, exhaustive small
joint enumeration, delayed/cross-checkpoint replay, full journal rare-opposition
regression, future forks and nine corrupted-artifact controls precede collection.
Reference QueryMode may omit unused branch calculations, not requested fields;
all mode-specific results must agree with the full reference call.

Reconstruct ALL 1,440 new-model arms independently: 3,456,000 issued packets,
6,912,000 clean/W1 scalar comparisons, requests/choices, snapshots, metrics and
cost sums. Stable ordering uses independently validated recorded scores to avoid
bitwise near-tie arithmetic demands. All 240 Full/Adaptive controls bitwise
unchanged excluding cost against V66 and V60; all 40 populations identical.
Readback streams worlds individually. Three serial benchmark repetitions.
Preserve numerical failure source/log before its repair. Freeze negative science
and performance results unchanged; no passing component closes a whole goal.
