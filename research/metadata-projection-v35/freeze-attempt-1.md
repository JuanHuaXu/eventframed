# Pre-Confirmation Setup Failure

2026-10-03. `node research/metadata-projection-v35-audit.mjs --freeze` exited1:

```text
evalmachine.<anonymous>:4
const core=sealed.slice(sealed.indexOf('const dir='),sealed.indexOf('
                                                                    ^
SyntaxError: Invalid or unexpected token
```

The extraction matched the text `function below` inside V34's nested extraction
string instead of the actual declaration. Local repair uses a newline-anchored
declaration boundary and checks start/end ordering. Sealed V23-V34 checkers and
all runtime/protocol/gates unchanged. No V35normal source freeze was written.

The subsequent run command correctly refused absent `freeze.json` (exit1/ENOENT).
No normal raw, load log or run manifest existed, and no normal trial started.
This is an auditor setup repair BEFORE prospectively freezing confirmation,
not retrospective adjustment of an experiment's results or acceptance gate.
