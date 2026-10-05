# Numerical and empirical requirements after V81

All seven WHOLE goals OPEN. This is a concrete lead, not a tested rescue.

The protected support preflight reproduces a positive-probability rejection at
reset0/T12800: independently logP=-1321.2095576498039, below binary64's smallest
positive representable probability. Distinguish genuine zero from underflow.
Do not add a floor, resurrect absent masks or reinterpret a numeric failure as
an impossible observation. Current source is frozen and its negative preserved.

[Blanchard, Higham and Higham](https://eprints.maths.manchester.ac.uk/2744/1/paper.pdf)
(January2020 manuscript, equations1.3/1.4 and section4) analyze max-shifted
log-sum-exp and softmax and support shifted normalization over unshifted
exponentials. This motivates a representation repair, not proof of this learner's
accuracy, posterior-approximation control or preserved rank traces.

Separate next fork: store component log masses persistently, add log likelihoods
and log transitions, use shifted log-sum-exp at normalization/retained masks.
Only exponentiate for point summaries; never overwrite log state with an
underflowed summary. Carry structural zero explicitly. Review tiny reset/member
probabilities, repeated min-subnormal rates, invalid infinities and sort ties.
Log-domain component mass alone may not protect member conditional rates; audit
the latter independently before claiming full numerical-domain validity.

Pending needs a declared LogProbability per branch. A finite log mass below
representable probability is not structural zero. Returning Probability=0 must
carry an explicit representational status/bound, or abstain honestly. A stable
normalizer followed by exp(logP) cannot solve this interface problem. Validate
log-domain branch mass/tower and updates against a closed-form long-history
oracle, ordinary dense oracle, exact support/mask identities and negative
impossible-pair controls. Preserve support epochs and forward/cache invariants.

Do not let numeric/performance work replace quality evidence. Read the COMPLETE
V81 frozen screen, including reset-only and rejected-packet arms. If promising,
integrate a separately frozen adapter on unchanged original population/native
Full/Adaptive controls and quality/non-harm/recovery gates. Include missingness,
interaction/subgroup generators, varied independent fitting samples, large
frontiers and full history cost. The small known-baseline screen supplies none
of those successes. Whole-loop timing, not query microseconds, governs resource
decisions. Frozen input replay and received-label audit must precede confirmation.

Useful external error-controlled splitting, untouched outcome-labeled agent
utility, durable mixed-write/epoch continuity and equal TOTAL acquisition cost
remain separate required work. No reserved seeds/private/sealed labels or
production/paper/publication access is authorized by these research notes.
