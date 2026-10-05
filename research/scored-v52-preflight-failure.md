# V52 generator preflight failure

Before any fixture generation or outcome experiment, the exclusive generator
stopped after local/hybrid/collector generation: its mandatory rename assertion
incorrectly expected the hybrid constructor name in the auditor source.
The following gofmt command reported missing not-yet-generated files. The
following test returned zero with NO TESTS TO RUN, so is NOT validation evidence.

Repair is confined to the generator: optional renames use replaceAll; formatted
existing outputs must equal the deterministic result byte-for-byte on resume.
No generated sources are overwritten, no prior sources changed, no seeds
consumed, no model, gate, sampling scope or scientific parameter altered.
The subsequent frozen run must contain positive executed tests and audits.

The next compile preflight caught retained V51 names on two new study helpers
(duplicate declarations) and a new test referring to `alpha` instead of the
model's actual `hazard` field. No outcomes had been generated. The generator's
explicit preflight-repair mode only rewrites its OWN prior hash-proven outputs;
all original source hashes remain checked. This mode is not used after freeze.

The harness generator also stopped before writing its outputs: replacing
`Fixture` before `FixtureChecks` consumed the latter prefix and tripped the
mandatory transformation assertion. Longer names are now transformed first.
Dependent syntax/run commands reported missing files and did not run a study.

The generated runner then failed JavaScript parsing before creating its root:
the regex end-anchor followed by a quote was interpreted as replaceAll's `$'`
replacement expansion. The generator now returns literal replacement text via
a callback. Both generated scripts must pass syntax checks before execution.
