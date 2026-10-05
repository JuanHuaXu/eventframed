# Predictive-risk acquisition

Frozen before scoring; consumed fixed-tape v120 research. Keep all2688 runs,
unchanged priors, experts, query times/pools and perfect unit-cost label service.
Only replace current-state entropy utility. No expert retraining or fresh claims.

Use eight most recent visible forecast vectors, origins[t-7,t], as equal-weight
virtual predictive probes. Under each current expert, a virtual outcome has the
archived probe probability. It is conditionally independent of the outstanding
label given the current expert; it is NOT the actual old outcome. This declares
a working decision model, not reuse of training outcomes or knowledge of future
inputs. Recent probes are a proxy whose distribution mismatch can cause failure.

For candidate label j, reuse exact hypothetical refilter outcome masses and
current-state conditional weights. Compute average squared movement of each
probe's posterior predictive mean, averaged over the two hypothetical labels.
By total variance this is the expected Brier Bayes-risk reduction in that virtual
model. Select maximum, earliest-origin tie; reveal at t+1 after issued forecast.
No Q or hidden Y in query selection. No immediate revision of served forecasts.

Compare against natural, random, entropy, old disagreement and exact current-state
information. All168 cells must protect lower paired gain>=-.01; delayed terminal
changing cases1,2,4,5,7,8,19,20 need mean>=.005 and lower>0 against all five.
Use mean +/-3.5SE over32 trajectories, exploratory not simultaneous. Preserve
all failures. Equal per-trajectory costs and immediate-schedule identity required.

Validate Bayes-risk formula against explicit label/expert/virtual-outcome sums,
zero value when experts predict identically, as-of poisoning and replay. Source
and reference hashes recorded. Reference cost adds O(8*K) per hypothetical query
to existing two refilters; slow path. Record offline timing, not serving claims.
