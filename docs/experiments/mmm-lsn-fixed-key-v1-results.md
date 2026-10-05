# Fixed-width availability key v1: finite contract pass

Protocol: [mmm-lsn-fixed-key-v1-protocol.md](mmm-lsn-fixed-key-v1-protocol.md).
A private collection stored a new `available_at_sort: StringField` derived
from `time.Time.UTC()` using exactly nine fractional digits. At the latest
durable LSN, SQL returned `{early}` at `.1Z`, `{early}` at `.12Z`, and
`{early, late}` at `.15Z`. All three results repeated after close/reopen.
A different time-zone representation of the same instant generated the
same sort key. This corrects the exact fractional counterexample that
falsified the existing variable-width `available_at` SQL filter.

This is only a finite **raw-storage contract** result. Rows were inserted
directly into a private collection; no eventframed batch version, motion,
sidecar, old-row migration or loaded latency was tested. The key is safe
only if derived atomically from the authoritative EventFrame availability
time, validated on every write, and present for every row a serving query
might consider. A partial migration must fail closed, not silently omit
old rows. Production remains unchanged.

```sh
EVENTFRAME_RUN_LSN_FIXED_KEY_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchLSNFixedKeyV1$' -count=1 -v -timeout 2m
```

Source SHA-256 at run: test `05241153e8be3b8e32a9d3dc1856c7e5e5806eb3ab24c1d522fbf475e8267f1d`;
protocol `815ea35195b731b39a294371eb3e8412101763ed2d92ca5352a20fed17f3cf5a`.
