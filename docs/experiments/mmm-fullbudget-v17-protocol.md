# Full-budget control v17

Frozen before evaluation. Compare subset64 adaptive retention with its original
confidence stopping against the same learner using the remaining six-coordinate
budget. Only arm2 loses the early confidence stop; all view ordering/value rules,
models, data, mixture updates and other arms remain unchanged. No true-rule
information enters either mode. This tests v16's stopping hypothesis, not a new
trained model or unrestricted exhaustive observation.

Use majority/multiplexer families, fair/biased/clustered inputs and the four v14
scenarios, six fits and two streams per fit, two splits and two modes=1152 streams.
Fresh fit2026110201, design2026110202, confirmation2026110203; unchanged seed
encoding. Six-coordinate cap in both modes; report actual mean costs, not just
the common cap. Some extra reads can worsen finite-sample estimates.

Require every full/post mean Brier harm<=.01 versus current stopping and versus
fixed counts. Require clustered majority shift128 post gain>=.005 versus current
stopping in both splits (the specific predeclared v16 hypothesis). Also retain
the existing every-family/generator shift128 post gain>=.005 versus fixed counts.
Report all failures and descriptive six-fit95% bootstrap intervals (20000
resamples, seed2026110299). No simultaneous-coverage or production claim.

Verify original-stop parity, unchanged arms0/1/3, full-budget cost6 on complete
synthetic readers, paired/delayed feedback, exact replay and source hashes.
Streaming raw artifacts. Passing quality alone cannot justify runtime adoption;
extra observation cost and real-task semantics remain separate requirements.
