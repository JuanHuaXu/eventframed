# Kalman origin-time update v3: frozen rescue fails

Date: 2026-10-02. The isolated [protocol](mmm-kalman-origin-v3-protocol.md)
**fails** in both design and untouched confirmation. Correcting the
linear-Gaussian working filter's delayed-label timestamp changes scored
predictions only negligibly on these fixed-delay streams. It does not
rescue Goal 2 or Goal 4. Production and the whitepaper are untouched.

## Audit

The [design tape](mmm-kalman-origin-v3-design.jsonl) contains 160
trajectories (SHA256 `a8977c058b7085a4f4ad55929640cda9c13d545c2f667597e65c329e9b165e79`);
the [confirmation tape](mmm-kalman-origin-v3-confirmation.jsonl) contains
160 different trajectories and independently refitted baselines (SHA256
`02e167fc8dd7d38b5e1387755359340156da752278899457a9871e76b3193b19`).
Each has 32 trajectories per scenario and 512 forecast-before-feedback
ticks per trajectory. Seeds, source hashes and fit offsets are embedded
in the manifests. The [independent verifier](../../research/kalman-origin-v3-verify.mjs)
replays every new forecast from recorded arrived packets, recomputes
all Brier windows, checks delivery order and zero-delay parity, and
evaluates the frozen paired gate. It returns `designPass=false` and
`confirmationPass=false` for both predeclared process-noise values.
Focused zero-delay/no-future Go tests pass under `-race`; `go vet`
passes for the package. Across the two splits, 19,188 and 19,164
arrived audited labels supplied updates, respectively.

## Scored outcome

Lower Brier is better. `origin gain` is delivery-time Brier minus
origin-time Brier over clocks 272..335, the first 64 clocks after a
changed label could arrive. Intervals are paired mean +/-3.5 SE over
32 trajectories; they are not simultaneous confidence sequences.

| Split | q | Delayed delivery -> origin Brier | Origin gain [lower, upper] | Last-64 Brier | Stable full Brier: base -> origin | Interaction tail Brier: last-64 -> origin |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Design | .0005 | .308651 -> .308650 | .000001 [-.000007,.000009] | .251336 | .062144 -> .064543 | .244214 -> .304946 |
| Confirmation | .0005 | .336954 -> .336953 | .000001 [-.000007,.000009] | .253320 | .063380 -> .067914 | .242689 -> .302646 |
| Design | .01 | .256657 -> .256628 | .000029 [-.000016,.000075] | .251336 | .062144 -> .079256 | .244214 -> .373497 |
| Confirmation | .01 | .294212 -> .294214 | -.000002 [-.000048,.000043] | .253320 | .063380 -> .083046 | .242689 -> .370183 |

The frozen delayed gain target was at least .01 with a positive lower
endpoint, plus no more than .005 Brier harm against last-64. Neither
candidate approaches the gain threshold. Low-q stable harm has paired
upper endpoint .00777 in design but .01085 in confirmation, over the
.01 ceiling; high-q upper endpoints are .02594 and .02819. Both
candidates also fail the nonlinear interaction guard by wide margins.
The zero-delay `shift128` forecasts match their same-q delivery-time
controls to within 1e-10 at every clock.

The frozen 272..335 window includes one still-blind forecast at 272:
feedback due then arrives *after* that forecast. A post-hoc check of
273..336 gives origin gains of .000001/.000002 for low q and
.000028/-.0000005 for high q in design/confirmation, respectively.
This does not change the failed gate; the frozen protocol retains its
original wording, and this result records the timing correction rather
than silently moving its threshold.

On the delayed stream's blind clocks 256..271, low-q Brier is
.43419/.43180 in design/confirmation and no changed audited outcome
has arrived; the onset-information result still applies. The origin
filter retains 1,072 bytes. Maximum measured per-stream forecast p99
was 291 ns and update p99 was 29.5 us, both isolated timings below
the 50 us component cap. These exclude fitting, storage, queues and
full serving latency.

## Interpretation

For this fixed-delay, in-order generator, updating at the origin time
instead of delivery time is mathematically cleaner but not a useful
recovery rescue. It cannot reveal a change before evidence arrives,
and the observed post-arrival deficit is dominated by model/selection
limitations rather than this covariance-timing approximation. The
next plausible Goal 2/4 lead is a distinct bounded challenger with
nonlinear interactions and an explicit stationary-protection gate,
tested on fresh generators and equal observed-label budgets. Do not
retune these consumed cohorts or promote either filter to serving.

Reproduce the audit with `node research/kalman-origin-v3-verify.mjs`.
