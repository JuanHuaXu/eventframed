# Lazy repair branch controls

Five explicit small graph fixtures exercise reconnectNeighborsOptimized directly:
dense, all level-zero links cleared, alternating level-zero links cleared, one
valid neighbor plus an absent ordinal, and already-canceled context. These are
constructed branch tests, not representative corpus or causal-inference data.

Control and candidate each passed under the race detector (1.283s and1.265s).
All initial/final graph-state hashes and error outcomes matched between overlays.

| Case | Graph changed | Control pair calls | Lazy pair calls |
| --- | --- | ---: | ---: |
| Dense | no | 28 | 0 |
| Sparse | yes | 28 | 28 |
| Mixed | yes | 28 | 28 |
| One valid neighbor | no | 0 | 0 |
| Canceled | no | 0 | 0 |

The sparse fixture explicitly asserts that reconnection changed state. Thus the
candidate does not achieve its reduction by suppressing all repair. The mixed
fixture tests delayed matrix creation after some neighbors already meet the
threshold. Cancellation is pre-operation, not an asynchronous cancel-race proof.

The repository's test-name inventory exposed only two existing deletion/repair
tests under the requested patterns, motivating these added controls. The new
cases complement, not replace, the prior deletion/reclamation race runs and
serial end-to-end mutation capture.

This strengthens branch correctness evidence. It does not erase the prior failed
timing screen, establish arbitrary concurrent equivalence, or implement private
transactional graph updates. The overlay remains research-only. All seven whole
research goals remain open.

Raw data and comparisons: `lazy-repair-cases/*-results.json` and
`lazy-repair-cases/comparison.json`; harness: `lazy-repair-cases-test.go.txt`.
