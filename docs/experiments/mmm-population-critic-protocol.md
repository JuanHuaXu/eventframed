# Population-target critic: frozen comparison

Hypothesis, not confirmed root cause: sampled future inputs add avoidable noise
to acquisition targets. Alternatives include insufficient observable state,
model misspecification, and inherently unpredictable future evidence. Keep the
four error-state factorial models, features, ridge .01, centering, pool weights,
tie tolerance and abstention threshold unchanged; change only training targets
from sample-input expected gain to integrated population expected gain.

Train on the672 phase0 delayed pools only. Evaluate all2688 consumed records,
reporting actual-answer sampled loss, both-answer sampled expected loss, and
both-answer population expected loss. Preserve all old controls. Compare forced
query to forced random/entropy; gated query to random/entropy on precisely the
same gate mask. Include previous sample-trained decisions on the same outcomes.

Apply the existing actual-answer advancement screen separately to each variant
and mode: every phase1 delayed cell has paired lower gain bound>=-.001 against
both controls, and both switching cases19/20 have positive lower bounds against
both. Also report population and sample-expectation screens, without substituting
them for the primary screen. Intervals are mean +/-3.5SE over32 trajectories,
descriptive only, not simultaneous or anytime guarantees. No winner selection,
hyperparameter tuning, untouched-confirmation claim or whole-goal success.

Verify source/target/projection identity, all controls, exact reproduction of the
old models under old targets, phase1 target poisoning, oracle-field isolation at
selection, complete-delivery identity, solver residuals and deterministic replay.
Measure fit and selection costs separately, excluding generation of existing
Bayesian and prequential features. Do not modify existing hashed artifacts or
production software. Negative results falsify this target-only rescue on this
screen, not all possible learned acquisition policies.
