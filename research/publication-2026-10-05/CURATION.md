# Curated Research Branch

The local checkpoint is a recovery archive, not the publication parent. The
curated branch is built from its original public parent, so excluded checkpoint
blobs cannot remain reachable through an ancestor commit.

## Findings And Actions

- CONFIRMED: 718 source/build metadata candidates screened. Machine-local records are excluded from
  publication; original hashes remain in the local archive.
- NEEDS INVESTIGATION: 161 copied overlays, generated build records, and archive
  entries. Excluded rather than assuming the project's Apache declaration
  relicenses copied dependency code. First-party control implementations remain.
- NEEDS INVESTIGATION: 37 result artifacts have source-text or identity-shaped
  fields. These are not asserted to be private leaks; their provenance and
  redistribution boundary needs review. Retain aggregate reports instead.
- CONFIRMED: 13 executable harnesses embed workstation paths and four alternate
  module files depend on local backend forks. Keep them in the local checkpoint,
  not in the portable public branch.
- CONFIRMED: three local checkpoint machinery/working-tree metadata files.
  Excluded. Forty-six reports have workstation prefixes replaced with a marker;
  numeric scientific content is not changed and original/curated hashes are
  recorded separately.
- CONFIRMED: `cohort-batch-v46-generation.json` is a portable 58-function
  source-equivalence descriptor, not machine-local build state. After complete
  schema/path review, it is restored verbatim because an existing regression
  test consumes it. The failed missing-descriptor test remains recorded locally;
  the test is not skipped or weakened. This is an explicit exception to the
  filename-only metadata exclusion, not an exemption for other manifests.
- RECOMMENDATION ONLY: future public dataset redistribution requires separate
  provenance/license review. Signature and schema scans do not establish the
  absence of arbitrary inferable PII or certify runtime security.
- CONFIRMED by the first portable compile: three legacy harnesses depend on
  unavailable backend-overlay methods; archived Go fragments also lack their
  original surrounding packages. Preserve research-directory Go snapshots and
  those three harnesses verbatim under `research/code-archive/` as `.go.txt`
  reference assets, rather than altering their algorithms or compiling them as
  active module packages. `reference-assets.json` maps every original path and
  hash. The original failed compile remains recorded locally.

The preexisting public parent is preserved. This audit covers the checkpoint
additions and new prospective baseline files, not a rewrite or certification of
all prior Git history. No original working-tree files or failed trials are
deleted. Historical source, frozen protocols, compact numeric evidence and
negative results remain available; raw/large/local-only artifacts do not become
passing studies simply because they were omitted from publication.

## Baseline Decision

Adaptive becomes the prospective research accuracy reference under the original
400-ms complete-core budget. Full remains the speed control. Eighteen recorded
same-cohort arms were compared; the two controls were freshly replicated on all
120 cells each, with 576,000 independently recomputed expected issued losses.
Adaptive mean/worst core was 197.731/246.690 ms; Full 49.285/73.614 ms. All pass
that synthetic-core budget. This is not loaded agent latency or untouched
confirmation. V83 retains its numerical-component role, not accuracy promotion.

Publication is a research branch, not a daemon deployment or merge into main.
No remote push is performed by this work package.
