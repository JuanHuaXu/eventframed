# Confirmed latent-weight underflow

Allowed static2expert/200trial model with prior(.5,.5), advice(1e-6,1-1e-6),
100 non-useful then100 useful labels predicts0.000001 rather than near0.5.
Linear posterior multiplication zeroes the second state during the first block;
zero then becomes absorbing even though all model likelihoods are positive.
This is a model arithmetic defect, not poor data or a quality-threshold failure.

This directory stores72 exact pre-repair source copies, freeze, failing race log
and command. The repair stores latent messages in log space and uses stable
log-sum-exp transitions/normalization. Rounded output weights never feed back.
Original advice and issued forecasts remain immutable. Scratch validation still
precedes publication; every lifecycle gate and configured cap is unchanged.
New source versions use new labels; no old result is overwritten.
