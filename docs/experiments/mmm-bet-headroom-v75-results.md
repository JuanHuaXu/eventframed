# Betting headroom v75: mechanism supported, runtime benefit untested

Deterministic known-law calculations support testing a predictable learned
betting rate. This is not a new detection experiment, an oracle bound on every
algorithm, or a change to v74's failed adoption decision.

Expected log-wealth growth per observation, using exact channel moments and
the favorable sign:

| Case | Fixed rate | Bounded oracle rate | Fixed growth | Oracle growth |
| --- | ---: | ---: | ---: | ---: |
| Sparse | 0.25 | 0.408889 | 0.058024 | 0.090133 |
| Homogeneous | 0.25 | 0.456101 | 0.047156 | 0.061129 |
| Negative | 0.25 | 0.408889 | 0.058024 | 0.090133 |
| Weak | 0.25 | 0.375510 | 0.011971 | 0.017575 |

For the opposite signs, the boundary null and zero-evidence control, the oracle
chooses rate0. All caps were computed over every channel and all possible
D in[-1,1], not just outcomes supported by the known law. Thus factors remain
at least.08 even if an outcome considered impossible by that law occurs.

The optimizer is checked against a10001-point grid for both signs in all six
cases. Probability mass, zero conditional mean of the control term and endpoint
factor bounds pass. All12 rows are retained in `mmm-bet-headroom-v75.json`,
including the non-favorable/null results. Source/protocol hashes are embedded.
Reproduce with `node research/bet-headroom-v75.mjs`.

This is an upper optimization result only for the declared one-step growth
objective, fixed q/m and allowed rates. Neither reciprocal growth nor its
percentage change is a demonstrated reduction in restricted mean delay.
Learning uncertainty, transient model errors, restart mixtures, missed deadlines
and finite evidence budgets still determine actual stopping performance.

Source basis: [Waudby-Smith & Ramdas](https://arxiv.org/abs/2010.09686), for
predictable betting and time-uniform inference. Our concrete optimization and
all-channel cap are separate derivations, not a verbatim implementation.

The next candidate can use the existing32-observation histories to estimate a
three-outcome distribution with prior counts(.5,1,.5) for(-1,0,+1). This matches
v74's existing smoothed first/second moments, so it needs no new hidden labels.
Choose the bounded rate from that pre-query model, never from the revealing
outcome. Allowing zero bets is important to test: it may avoid wealth decay
during uninformative periods, but may also delay response while the model learns.
Freeze and test against v74 on new seeds before claiming an improvement.

No detector or serving code changed. All research directions remain open;
there was no commit, push, production access or whitepaper modification.
