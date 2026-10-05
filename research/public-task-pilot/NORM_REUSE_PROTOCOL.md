# Immutable vector norm reuse screen

Freeze before candidate execution. Store squared norm computed during existing
PrepareLayered validation; never trust a caller-supplied cached value. Keep the
same float64 summation order and sqrt(aa*bb) denominator, not pre-normalization
or inverse-norm multiplication. Reuse query norm already computed by search.

Require bit-identical scalar cosine against the original for finite nonzero
vectors and unchanged graph capture results. Record field storage overhead and
all allocation changes. No unbounded metric cache or mutable shared cache.

Performance screen uses four existing insertion benchmark cases, one second,
three repetitions per arm, CPU4, serial arms: archived control then candidate.
Require candidate median ns/op <=90% of control for every case and <=105% bytes/op.
This ordered screen is diagnostic; even a pass needs fresh alternating-order
confirmation and the original sustained-load gate. Retain failures unchanged.
