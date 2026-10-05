# Uncertain renewal inference: fixed-trace diagnostic

Freeze before scoring the consumed v11 traces. This isolates inference on
identical acquired evidence; it is NOT a closed-loop acquisition experiment.
The old planner chose every action. Do not claim the new posterior would make
the same choices or spend credits the same way under replanning.

For each type and hypothesis, introduce independent .5-prior binary ordinary
mode O and renewal mode F, plus root R~Bernoulli(q_type,h). The first ordinary
report is R; later ordinary reports equal R if O=0 and are iid q if O=1.
Renewal reports equal R if F=0 and are iid q if F=1. Keep one joint posterior
over (O,F,R) per type conditional on h, with shared h posterior. Crucially, R
exists before either channel is queried, so a renewal can precede the first
ordinary report. No learned trust threshold or revised prior is fitted.

Process all640 prior episodes and all four acquired traces. Verify the
conditional local-state recursion against direct batch marginalization over
latent modes/root, including both channel orderings and conflicting reports.
With F prior1, recover the old certain-fresh model. Freshness posterior is
model-conditional, not an external authenticity certificate.

Primary fixed-trace screen concerns the old mixed-planner traces. For each
split require false-renewal final-Brier improvement >=.02 with paired lower>0;
on independent20/copied20/mixed20/copied05 require paired lower gain>=-.01.
Report means +/-3.3 SE over64 trajectories; these are descriptive, not
simultaneous. Keep both splits and every trace. Report confidently-wrong counts,
final and credit-area Brier, and freshness posterior. Other planner traces are
matched inference controls, not alternatives selected after seeing results.

Success here would justify a separately frozen closed-loop test, not adoption.
Failure can be caused by overcautious treatment of genuine renewals, incorrect
prior structure or remaining misspecification. Copies from a root other than
the declared ordinary root are outside this model, so even a passing result
would not cover all evidence-poisoning patterns.
