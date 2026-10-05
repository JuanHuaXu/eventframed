# Batch normalization under changed-work service load

Repeat the changing-work fixture with off/direct/batch arms instead of a memo
arm. Neither fitter caches results. Both use the same transform and context
checks; only family normalization changes. Four-group and64-distinct histories,
three rotated arm orders each,64 recalls/four readers,16 future writes,
GOMAXPROCS4, actual persistent service, temporal compatibility,100ms/capacity16.

Each successful result must match the precomputed pairwise-reference forecast
for its own history to1e-12. References are verification only, outside timing.
Record all offered-job outcomes; no false completion credit for canceled fits.

Frozen screen: all arms error-free with overlap and terminal accounting;
batch completion>=52/64 and >=direct in every cell; batch p99<=1.10*off and
<=1.10*direct. No cache hits permitted. Preserve failures. This does not measure
semantic learner usefulness or establish population tail latency.
