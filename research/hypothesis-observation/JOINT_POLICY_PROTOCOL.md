# Joint forecast/acquisition risk game

Design on the consumed80-world grid, not fresh empirical confirmation. Keep
the entire465-row screen, observation contract, six-query budget and controls
from GLOBAL_POLICY_PROTOCOL.md. Change the best-response class to permit both
observation choices and forecasts to depend on observed sufficient state.

For each adversary weight vector, combine per-world unnormalized outcome masses
at state s into v_y(s). The Brier-minimizing forecast is p_y=v_y/sum(v), with
uniform p when the mass is zero. This follows from the excess-risk identity
L(p)-L(p*)=sum(v)*||p-p*||^2. Feed the minimized state loss into the unchanged
backward observation oracle. Final and area row weights can induce different
forecast rules at terminal and pre-query stages. This is a decision-theoretic
predictor, not a claim of one ordinary sequential Bayesian posterior.

Use the same128 adversary rounds, learning rate32 and uniform mixture of
complete policies sampled before the root observations. Do not average forecasts
across incompatible histories or select a policy using realized truth. This is
a mixture of policy/forecast pairs with expected score equal to the mean of
their scores. Each rule sees counts/root only, not actual regime or target.

Report primal/dual bounds and original pass counts. A negative upper bound
would demonstrate finite-screen feasibility, not general learning or empirical
validation. A positive lower bound would suggest even this joint class fails.
Floating-point bounds are not outward-rounded certificates. Do not assume
convergence or promote a consumed-grid fit. Save weights and risks for replay.

Verify forecast simplex, the Brier excess-risk identity, zero-mass handling,
tiny exhaustive observation oracles, weighted forward/backward agreement,
archived-control domination under each oracle objective, replay and mixture
reconstruction. New independent noise/mechanism tests are required after any
promising result. Production and all archived artifacts stay unchanged.
