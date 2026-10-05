# Post-run closure-auditor preflight

The original Go-list template attempt failed before writing any audit artifact:
`template: main:1: function "json" not defined`. This installed Go-list tool does
not supply the assumed JSON template helper. Preserve the failed attempt; no
experiment source, prediction, seed, gate, parameter or metric is changed.

The replacement invokes `go list -deps -test -json` and uses Go's standard
streaming JSON decoder to produce an array for Node's standard JSON parser.
This is an independent post-run provenance check, not a prospectively frozen
experimental source and not an inherited runtime-toolchain guarantee.

The first closure assertion identified nine missing calibration files, but six
were dependency-test metadata and not actual compiler inputs. Go-list test
variants already expand compiled test sources in GoFiles. The corrected check
counts GoFiles/CgoFiles/EmbedFiles only. Any REAL omission is reported with
post-run hashes and a false prospective-completeness flag, not silently added
to the old freeze. The original assertion failure remains part of this record.
