# Prequential reliability v20 diagnostic

Frozen before replay of consumed v19 mixture_stop traces. Not a new policy or
confirmation. For each predicted class, retain last64 delivered outcomes from
confidence-stopped forecasts. Journal whether prediction was correct and its
nominal correctness probability max(p,1-p). Before receiving the current label,
warn when at least32 such outcomes exist and (correct+1)/(n+2) is more than .03
below their mean nominal correctness probability. Otherwise mark unknown support
or no warning. Beta(1,1) smoothing is a working estimate, not a stationary-law
guarantee or a confidence sequence. No peeking at undelivered labels.

Measure warning coverage, error capture and warned/nonwarned error rates on
current confidence stops, by family/generator/scenario/split. This asks whether
past reliability identifies problematic frames, not whether extra reads help.
Independent fresh-policy tests would still be required before adoption.

Go/no-go diagnostic: for clustered majority shift128 confirmation, warn on<=50%
of confident stops, capture>=50% of their errors, and exhibit warned error rate
at least twice the nonwarned rate. Predeclare failures rather than tune bins or
thresholds afterward. Also preserve stationary results and delayed arrival checks.
No falsified calibration certificate is inferred from an empirical warning.
