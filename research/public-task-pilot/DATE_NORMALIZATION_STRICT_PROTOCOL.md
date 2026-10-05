# Strict normalization follow-up

The first normalization prototype passed calendar equivalence but a subsequent
boundary probe confirmed malformed ISO prefixes were accepted:2004-03-020,
2004-03-02suffix and2004-03-02+01:00 all became2 March2004. This violates the
standalone-token contract. Preserve the original implementation/artifact; it is
not approved. New strict variant rejects unsupported adjacent characters and
time suffixes before invoking the unchanged calendar normalizer. Conservative
rejection can reduce coverage; it must not be called general date parsing.

Before strict replay: require boundary regressions, the same full-cycle calendar
test against the strict module, and exact equality of all saved public-case
results with the prior normalization run. Do not alter query rules, scores or
laws. This tests a repair on consumed fixtures, not fresh evidence of accuracy.
