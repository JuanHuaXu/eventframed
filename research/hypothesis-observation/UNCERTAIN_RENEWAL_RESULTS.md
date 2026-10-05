# Uncertain renewal: confidence rescue, not overall rescue

Fixed-trace diagnostic under [the protocol](UNCERTAIN_RENEWAL_PROTOCOL.md).
All640 consumed v11 episodes and all four original acquisition policies were
retained. Only inference changes; no action was reselected. The primary screen
FAILS:4/10 required gates pass.

## What the model changes

Each type carries a joint conditional posterior over ordinary independence,
renewal independence, and their shared first-report root. A renewal-labelled
report can be iid or a copy of that root. The root exists even if renewal is
queried first. This is not just a scalar trust multiplier; it changes the joint
likelihood and prevents repeated copies from automatically becoming independent
confirmations. Independence prior is fixed .5, without outcome-based tuning.

## Same-evidence comparison

Original split1 mixed-planner traces:

| Case | Certain-fresh Brier | Uncertain Brier | Old confidently wrong | New confidently wrong |
| --- | ---: | ---: | ---: | ---: |
| Independent20 | 0.129272 | 0.127063 | 2/64 | 0/64 |
| Copied20 | 0.210300 | 0.224100 | 2/64 | 1/64 |
| Mixed20 | 0.067544 | 0.096254 | 0/64 | 0/64 |
| Copied05 | 0.00000216 | 0.00002257 | 0/64 | 0/64 |
| False renewal20 | 0.744449 | 0.599900 | 24/64 | 0/64 |

False-renewal Brier gain is0.144549 [0.038058,0.251039] in split1. Accuracy
remains62.5%: it reduces unwarranted certainty, not classification mistakes.
Zero observed confidently wrong events is not a zero population rate; a95%
Wilson upper bound for0/64 is about5.66%. Brier remains worse than the v11
regular-only policy's0.492846, so this is not a full poisoning rescue either.

The cost is visible on genuine renewals: copied20 gain is-0.013800
[-0.026413,-0.001187], and mixed20 is-0.028710 [-0.050013,-0.007407].
Six protection gates fail across the two splits; four false-renewal/low-noise
gates pass. Both independent20 protection intervals cross the-.01 threshold,
so those failures are uncertainty failures, not proof of mean harm in each split.
All intervals are descriptive paired mean +/-3.3 SE over64 episodes, not
simultaneous or anytime-valid coverage. This is consumed-data diagnosis.

## Verification

Four component tests pass: all128 six-outcome sequences in two channel orders
agree with batch latent-state marginalization at every prefix; F-prior1 matches
the original certain-fresh model; contradictory root/renewal evidence rules out
copying under the declared model; duplicate slots are rejected. The root-first
ordering test also covers renewal-before-original, avoiding hindsight root use.

The [independent full-data verifier](verify-uncertain-renewal.mjs) directly
enumerates latent modes and root likelihoods from each acquired prefix. It
reconstructs35,894 forecasts and all final/credit-area Brier scores, with maximum
forecast discrepancy2.546e-14. It verifies input and source hashes. These checks
validate the implemented finite model, not its real-world adequacy.
All640 episode results, including every rescored forecast and summary, reproduce
exactly on full replay.

Artifacts: [all forecasts and source hashes](uncertain-renewal-v12.json),
[summary and all gates](uncertain-renewal-v12-summary.json),
[independent reconstruction](uncertain-renewal-v12-reference.json).

## Remaining lead

Fixed old actions may be poor for the new posterior. A separate closed-loop
experiment should let the model acquire reports that distinguish copied from
independent renewal, retaining the same cost and all genuine/false cases. It
must not assume improved choice will remove this tradeoff. Copies of an
unmodeled root, correlated sensor errors, and outcome-dependent source selection
remain outside the finite family and need explicit falsifiers. No learned
freshness posterior is an external authentication certificate.

No production or whitepaper changes. All seven directions remain open.
