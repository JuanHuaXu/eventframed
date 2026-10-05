# Cadence diagnostic contract checks

The race run in mmm-refit-cadence-contracts.txt passed in48.399s package time.
TestRefitCadenceContracts itself took46.88s. Four fixtures cover stationary case0
and changed case20, with complete and delayed/missing feedback. The32-clock
control matches the original five forecast components. The faster path preserves
the first8 forecasts, has exactly32 fit bundles, and buys no labels.

All128 fast-fit bundles have the full expected as-of origin lists. Eight poisoned
prefixes flip unarrived/current/future labels and all Q values without changing
issued forecasts. Input serialization is unchanged after the tests.

The independent JavaScript scorer passed four constant-forecast fixtures spanning
delay0/31 and missing/no-missing evidence. Brier is exactly0.25 and12 corrupted
fit clocks, origins or mixture forecasts are rejected. Full scoring also checks
every original and faster served mixture by direct probability recursion, both
full/terminal losses, paired trajectory signatures, and three mutation probes.
This is not an independent reimplementation of the underlying expert fitters.

The pre-cadence acquisition helper is archived as
research/acquisition-training-v1-source/acquisition_train_pre_cadence.go.txt.
Its SHA256380bc5bf91c8187a45721f2825f3c0d4cb25acb961e67db90ccf4c3dc378360a
matches the prior query-burst raw artifact's helper hash. Earlier results remain
reproducible despite the new explicit cadence parameter; its default stays32.

These checks establish implementation contracts, not predictive improvement.
The extra compute remains explicit:8 versus32 four-expert fit bundles per
trajectory. See mmm-refit-cadence-protocol.md for frozen analysis and boundaries.
