# Experiment artifacts

This directory versions frozen protocols, aggregate reports, calibration
artifacts, and compact machine-readable results. Raw `dataset.json`,
`*-dataset.json`, and session-derived replay corpora remain local and are ignored
by Git because they are large and may contain private conversation material.

Experiment reports must identify their dataset provenance and must not describe
a local raw corpus as a publicly reproducible artifact. Synthetic datasets that
need publication should be exported deliberately to a separate, reviewed
artifact bundle rather than added implicitly during an RC build.
