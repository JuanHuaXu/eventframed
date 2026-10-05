import fs from 'node:fs';
const source = fs.readFileSync('internal/observationgate/delayed_learning_test.go', 'utf8');
let body = source.slice(source.indexOf('func delayedLearningRun('), source.indexOf('func TestDelayedLearningContracts('));
function replace(old, next) {
  if (body.split(old).length !== 2) throw Error(`nonunique hook ${old}`);
  body = body.replace(old, next);
}
replace('func delayedLearningRun(', 'func delayedLearningDiagnosticRun(');
replace('schedule feedbackSchedule) (delayedRecord, error)', 'schedule feedbackSchedule, diagnostic *delayDiagnostic) (delayedRecord, error)');
replace('\t\tfor i, p := range predictions {', '\t\tif e := diagnostic.capture(step, name, y, g[1].entries[step%forecastJournalCapacity].forecast); e != nil { return r, e }\n\t\tfor i, p := range predictions {');
replace('\t\t\tr.Received++', '\t\t\tif e := diagnostic.feedback(step, p.Origin, statuses[1]); e != nil { return r, e }\n\t\t\tr.Received++');
replace('\t\t\t\t\tmodels.Version++', '\t\t\t\t\tdiagnostic.fit(step, audits[max(0, n-64):])\n\t\t\t\t\tmodels.Version++');
const imports = `// Mechanically derived from the frozen v85 driver; only read-only diagnostic hooks differ.
package observationgate
import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "math/rand"
 "sort"
 "github.com/JuanHuaXu/eventframed/internal/bayes"
 "github.com/JuanHuaXu/eventframed/internal/observation"
 "github.com/JuanHuaXu/eventframed/internal/observationexperiment"
 "github.com/JuanHuaXu/eventframed/internal/observationpreserved"
 "github.com/JuanHuaXu/eventframed/internal/observationrescue"
)
`;
fs.writeFileSync('internal/observationgate/delay_diagnostic_run_test.go', imports + body, { flag: 'wx' });
