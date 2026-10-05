# V63 journal equivalence and memory protocol

Frozen before collection. This is an isolated storage/ownership experiment,
not a scientific-quality rescue or production implementation.

The bank owns one raw receipt journal and position index. Static model tables
are shared; posterior, support/count state, windows, clocks and pending/member
counters remain separate. Journal forecast metadata records the bank-issued
mixture. Each selector row retains the independently issued expert joint kernels.
All dependent updates prepare before any publication; only bank mutations may
write the journal. Standalone Model mutations reject bank-owned children.

Gates: unchanged V62 public laws and complete request/receipt metadata,
independent top-down reconstruction, delayed/missing/expired receipts, future
forks, child authority rejection, fault atomicity, old epoch/foreign tickets,
readonly nonmutation. Keep existing 8,388,608-byte constructor allocation cap
at BOTH 150 and 200 members, windows 600/1200/2400 and depth7. No frontier shrink.

Race, vet, allocation and three benchmark repetitions use frozen/copy-verified
compiler closure and parent checkpoint. Race success is not concurrent API
support. Allocation is TotalAlloc, not RSS. The 2400-frame core loop includes
three models, selector,400 fixed paired packets and snapshots but excludes
adaptive nomination, delay queues, backend/RPC/persistence or loaded serving.
The original 400ms scientific full-loop gate is NOT inferred from this loop.

All seven whole goals stay OPEN until their full evidence exists. Subsequent
quality/recovery study must score the actual mixture and charge nomination and
delayed lifecycle costs. Existing V60 failures, consumed cohorts and unopened
confirmation remain unchanged; no production/private/paper/publication changes.
