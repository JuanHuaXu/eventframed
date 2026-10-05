# Fixed-observer birth attribution

Replay the complete160-trajectory local-birth archive; no new sample or tuning.
Keep its four subset arms and count controls exactly unchanged. Add two arms
that use birth weights for forecasting but retain the inherited-weight arm's
actual observation mask, values, guide and charged acquisition cost for the
corresponding gate. They cannot read any extra coordinate. Update their own
mixture histories only after their issued forecasts. Birth prior remains0.1;
authorization and local-model availability remain required.

Compare fixed-observer birth against inherited and coupled birth at the same
gate. Verify original metrics/tapes/fits byte-equivalently. Save fixed-arm tapes
and metrics, control archive hash and diagnostic sources. The .005 gain and
.01 non-harm screens stay unchanged, with mean +/-3.5SE descriptive intervals.
These are consumed exploratory data, not fresh confirmation or rare-error
coverage. No daemon, production, whitepaper or remote changes.
