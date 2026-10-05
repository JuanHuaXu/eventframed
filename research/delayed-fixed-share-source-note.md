# Research lead: delayed expert switching

Primary sources inspected on 2026-09-21:

- Herbster and Warmuth, [Tracking the Best Expert](https://mwarmuth.bitbucket.io/pubs/J39.pdf)
  (1998), Figure1 and Section5. A share step restores weight to previously
  unsuccessful experts. The original off-diagonal sharing convention differs
  from prior-reset sharing; do not silently reuse the same numerical alpha.
- Korotin, V'yugin and Burnaev,
  [Adaptive Hedging under Delayed Feedback](https://arxiv.org/html/1902.10433)
  (2019), Sections2,4.1,4.2; Algorithm4. The latent expert follows a transition
  `(1-alpha_t)*identity + alpha_t*p0`. Arriving old losses require recomputing
  forward weights from the earliest affected origin, not just multiplying
  today's weight. Unobserved losses contribute no loss factor. Their bounds
  contain delay penalties; the stated setup assumes losses are ultimately
  revealed by the horizon. Permanent missingness in our tapes does not meet
  that assumption. Their loss is weighted expert loss; squared loss of a
  convex probability mixture is no larger by convexity, but our budget guard
  changes weights and cannot simply inherit their regret theorem.

Proposed, not implemented: if continuous static-expert weights retain obsolete
preferences, compare a frozen switching model against unchanged continuous
and reset controls. Validate on small expert-path enumeration before tapes;
late evidence must revise filtering at its origin. Retain permanent missing
observations as absent factors, and label guarantees accordingly. No switching
rate should be selected from whichever value best fixes the current failures.
This source note does not alter the running continuous experiment.
