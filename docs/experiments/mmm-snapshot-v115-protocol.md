# Frozen snapshot-specialist v115 quality comparison

Freeze before generation. No parameter sweep, early outcome inspection, optional
sample extension, best-rate selection or threshold relaxation. Keep failed
candidates and all earlier protocols intact.

## Design

Seed base2160111500 plus phase*1,000,000 +case*10,000 +index*10 and roles0..4.
Audit effective seeds modulo2147483647 against every archived learner/null
block, including v113. Both phases use32 indices for each of12 existing cases;
each latent trajectory has immediate and delayed/missing schedules:768 latent
trajectories,1,536 runs. Phase-disjoint rule pools and all previous noise/input
conventions remain. This is still one Boolean fixture family, not independent
generator or prospective real-task validation.

Keep16 initial labels,256 scored frames, switch128, as-of64/32 fits every32
frames and six-coordinate observation cap. Delays0..31 and missing probability
.2 are frozen. Release earlier eligible labels before publication; release the
current zero-delay label after forecasting. Flush clocks256..287. Full audit
inputs used for fitting remain separately accounted; no extra labels per arm.

## Eleven arms

0 generic64;1 conservative prefix;2 full-role prefix;3 arrival Brier;
4 Brier-neutral;5 arrival log/no-neutral;6 arrival log/neutral;
7 fixed-rate delayed Markov;8 eleven-rate hazard mixture;
9 four-role delayed Markov with the new12800-budget gate/controller;
10 bounded immutable snapshot specialist, the sole candidate.

Arms0..8 remain unchanged and are checked against v113 on consumed seeds.
Arm9's6400-budget specialization matches archived Markov paths, probabilities
within1e-12 and accounting in immediate/delayed tests. Arm9 uses12800 here,
matching arm10's declared total gate allowance. Its four roles still evolve on
publication; it does not receive snapshot birth/retirement transitions. It gets
fresh gate IDs for each new fit and uses the same detached-publication rule.

Arm10 uses exactly two four-model banks, immutable generation IDs, initial role
prior(.95,.05/3,.05/3,.05/3), incoming rho=1/(v+1), full retirement-to-incoming
transition, and .001 fixed share across active banks. It does not tune hazard
rates or use oracle resets. Preserve same-ID rejection across publication and
remove credits to retired recipient IDs. Gate tests have32 predeclared starts,
neutral prior.5 and remaining.5 across other active models. Boundary12800
reserves8 windows*8 slots*2 views/.01; never lower it mid-run.

Advice uses identity-bound issued probabilities and exact bounded delayed
refiltering. Gate evidence drains in origin order and remains window-owned;
stale-window outcomes are advice-only. Neither non-rejection nor prior carried
confidence is a certificate of truth. These controls isolate budget effects but
do not make the four-role and eight-snapshot hypothesis families identical.

## Mandatory quality gates

All1,018 candidate10 gates must pass:

-960 non-harm:10 controls *2 phases *12 cases *2 schedules *2 segments
  (all256 and late128). Upper paired Brier-harm bound <=.01.
-58 gain:20 against generic (all-frame parity3/parity4/complement4 and late
  switches, both schedules/phases);6 against conservative prefix (delayed late
  parity4 and switches, both phases);4 each against controls2..9 (both delayed
  late switches in both phases). Mean gain >=.005 and lower bound >0.

Use paired mean +/-3.5SE over32 latent trajectories per cell. These are
approximate fixed-sample screens, not confidence sequences or a guarantee across
the full adaptive research history. Reusing the paired trajectory across two
schedules does not create two independent trajectories. Preserve all818 v113
requirements and add protections/gains against its candidate and the matched
budget control. Expected Brier is primary; expected accuracy and log scores are
secondary descriptions, never replacement pass criteria.

## Audit and reporting

Generate once to an exclusive0600 artifact containing every issued forecast,
mask, cost, fit origin set, feedback clock/accounting and source hashes. Check
all58 frozen sources. Independently reconstruct metrics, as-of fits, clock and
censor accounting, seed uniqueness and paired latent streams. Replay the whole
artifact before any promotion claim. Publish no intermediate winner selection.

Measure the whole eleven-arm fixture separately on a previously consumed design
trajectory, retaining every repeat. It includes fits but is not serving latency.
The earlier complete-journal benchmark already records roughly9.4-9.7ms immediate
and17.7-17.8ms delay31 for256 candidate frames excluding fitting/I/O, with about
10MB allocated by detached publications. Any quality result must retain this
cost context. No production, private-data, repository publication or whitepaper
change is authorized by this protocol. All seven roadmap directions remain open
unless evidence at each direction's own scope establishes otherwise.
