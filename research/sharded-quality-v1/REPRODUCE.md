# Reproduce

Saved-evidence verification needs Node and gofmt:

```sh
node research/sharded-quality-v1/evaluate.mjs
node --test research/sharded-quality-v1/evaluate.test.mjs
node research/sharded-quality-v1/verify.mjs
```

The verifier independently recomputes all means, regret, packed deficits, all
common-ID laws, phase comparisons and paired screens from the full Go transcript.
Oversize plain transcripts are retained locally and losslessly gzip-archived;
ARCHIVES.json binds both byte hashes. Verification accepts the .gz when the
plain file is absent, checking DECOMPRESSED bytes against the original run hash.
No measurement is removed to fit repository limits. gzip payloads are public
DESIGN numerical traces, not private sessions or native profiler captures.

Fresh collection: restore the f3231fa source and public_capture_load fixture
from the preceding public checkpoint, apply dependency-v3.patch to a PRIVATE
copy of libravdb v1.6.13, restore pinned lexer v0.1.12, and prepare matching
control/candidate copies with the candidate-only sharding option. The sibling
sharded-capture-load-v1 manifest/template/protocol pins the inherited V3 inputs.
Relocate absolute scratch paths in a NEW preparation, not historical evidence.
For a new output directory use prepare.mjs then run.mjs. Their overwrite guards
deliberately reject reusing an existing measurement manifest. Do not change the
installed module cache or start production daemons to reproduce the pilot.

Follow-up: learned laws, native contract nomination, larger heterogeneous corpus,
future-content perturbation, mutation/crash/cross-epoch recovery, full-load race,
sustained open-loop load/RSS and untouched outcome-labeled tasks remain required.
All seven original research goals remain OPEN regardless of this component pass.
