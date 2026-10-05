# Exact completed-fit reuse pilot

Retain prior persistent service fixture:64 recalls,16 future-dated writes,
four readers, GOMAXPROCS4, capacity16,100ms age. Both arms use temporal reuse.
Three alternating pairs compare direct fitting against a cold one-entry exact
work cache. Never prewarm the cache. Same fixed272-frame labelled workload.
Reuse is calculation reuse, not extra evidence or independent learning.

Key includes tenant,epoch,full history including arrival times,length,left/clock,
label cap,hazard,family mass. Store only full512 forecasts from successful fits.
The test-only implementation is tied to one algorithm version/process lifetime.
Each request retains scheduler deadline/snapshot checks. Changed-work unit
controls require misses for tenant,epoch,label,input,arrival,cap,hazard,mass and
clock/length. Query selects a forecast but is not new fitting evidence.

Screen: zero errors, overlap, terminal accounting; memo completion>=52/64 in
every pair and no less than control; memo p99<=1.10*control. This repeated-work
fixture is favorable to exact reuse, not evidence of real task cache-hit rate.
Do not extrapolate to differing histories or changing labels. Record failures.
