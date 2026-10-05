# Arrival-time role selector: component checkpoint

Subsequent quality result: [v106](../docs/experiments/mmm-arrival-v106-results.md)
FAILS16/318 gates despite exact replay. The lifecycle result below remains valid,
but it did not establish the required delayed recovery improvement.

Historical status at this checkpoint: component tests PASS; quality was not yet tested. The
[v105 timing diagnosis](../docs/experiments/mmm-origin-v105-results.md) motivates
removing about13 ticks of unnecessary prefix waiting before deciding whether
to discount the original losses themselves.

## One changed mechanism

The selector now applies each captured raw forecast's loss when its label
arrives. The version-scoped evidence gate still processes the original resolved
origin prefix, and still excludes forecasts from replaced versions. No priors,
learning rate, fixed-share rate, model family or acquisition policy is changed.
Fixed share still advances per delivered label; event-clock forgetting is not
silently included in this ablation.

Out-of-order arrival changes selector update order. With fixed share this is
not generally the same result as origin ordering. It is delayed expert-advice
learning, not an ordinary fixed-order Bayesian posterior, and no earlier
regret or test-martingale theorem is inherited automatically.

The prototype reuses the frozen journal transition, discards its temporary
selector update, and applies only the newly arrived loss. Gate computation
uses captured raw forecasts, not those temporary selector weights. When expiry
unblocks previously delivered labels, the gate advances but their selector
losses are not applied again. This deliberately preserves the tested gate
transition; optimizing away its temporary selector work is deferred until
quality justifies this variant.

## Verification

- Complete256-step immediate-feedback forecast and state parity with role carry.
- Literal arrival-order updates from captured original forecasts.
- Gate states and prefix accounting match the frozen journal transition.
- A later label updates the selector while the gate still waits for the head.
- Expiry and publication cannot replay selector losses or feed stale evidence
  into current-version tests.
- Duplicate, expired, reentrant and failed-reader paths do not mutate state.
- Race-enabled tests and vet pass.

`received` counts arrival-applied selector losses separately from the journal's
prefix-settlement counters. It must not be inferred from `Applied+BankOnly`
while delivered records remain buffered. The implementation retains the bounded
64-entry journal and256-issue research horizon; it is not a production API.

## Cost

[Paired output](arrival-routing-component-benchmarks.txt), Apple M4 darwin/arm64,
benchmark suffix -10, three500ms repetitions after trace replay and tests:

| Full256-frame fixed-model lifetime, delay8 | Time | Bytes | Allocations |
| --- | --- | ---: | ---: |
| Arrival selector | 1.377-1.405 ms | 135,280 | 1,215 |
| Origin-prefix role carry | 1.268-1.269 ms | 135,280 | 1,215 |

Approximately9-11% more component work on this fixture, with unchanged measured
allocations. These are not per-request latencies. Fitting, publication copies,
retrieval, storage and network serving are excluded. There is no demonstrated
quality payoff yet.

Source SHA-256:

```text
9dce725ca8d18970aa86909267354d9d27c628d96ca2f11e1eee7b69c696e6a3  arrival_routed_journal.go
c439590335679ffda5d491d5dc665419b53b7e6c71ec4d9f67393a49153977d3  arrival_routed_journal_test.go
```

Next: freeze a fresh paired experiment retaining generic, conservative and
origin-prefix carry controls. Test both change directions and protect stable
parity. Do not weaken v104's frozen claims or count this lifecycle pass as a
resolution of its12 failed quality gates. All seven directions remain open.
