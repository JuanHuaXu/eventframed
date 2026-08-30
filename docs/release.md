# RC0 release contract

Version: `0.1.0-rc.0`

RC0 is a reproducible source and test checkpoint for a local, single-user
OpenClaw memory pilot. It is not a production-readiness claim and does not
promote selective Bayesian inference, predictive snapping, or bounded agency.
Agency remains disabled by default.

## Release gate

From a clean checkout with Go 1.25.7 and Node.js 22:

```sh
cd plugin
npm ci --ignore-scripts
cd ..
make check build VERSION=0.1.0-rc.0
```

The Go binary is built with path trimming, a fixed version, no VCS-dependent
metadata, and an empty Go build ID. The TypeScript adapter is built from the
committed lockfile. CI runs the same race, vet, type, unit, and build gates.

Raw session-derived replay corpora are intentionally excluded from Git. Frozen
protocols, aggregate reports, and compact evidence remain versioned. A report
that depends on a local raw corpus is evidence for that recorded run, not a
self-contained public reproduction artifact.

## LibraVDB compatibility

| Component | RC0 status |
| --- | --- |
| Embedded Go module `github.com/xDarkicex/libravdb` v1.6.13 | Pinned by `go.mod` and `go.sum`; covered by the Go race suite |
| External LibraVDB contract server 1.9.36-beta.5 | Evaluation only; contract integration was exercised, but a dirty-journal shutdown panic was observed under replay load |
| OpenClaw 2026.7.1-2 | Development SDK pinned by `plugin/package-lock.json` |
| OpenClaw plugin API >=2026.6.2 | Declared runtime compatibility floor |

No external LibraVDB daemon is production-approved by RC0. A release candidate
that enables the remote contract path must pin a sidecar build that survives the
restart, deletion, replay, and concurrent-load fault matrix.

The RC0 npm audit reports no shipped production dependency vulnerabilities. It
does report advisories in the pinned OpenClaw development SDK tree used for type
checking. CI installs that tree with lifecycle scripts disabled. Those
development advisories remain a build-environment hardening item before a
production release.

## Supported pilot boundary

- local mode-0600 Unix socket
- one configured OpenClaw tenant
- recall and capture with explicit untrusted-context formatting
- agency disabled
- corrections may be measured in shadow mode

TCP exposure, multi-tenant identity, unattended agency, and general production
operation remain outside this release contract.
