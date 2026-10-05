import fs from 'node:fs';
const read=p=>fs.readFileSync(p,'utf8');
const write=(p,s)=>fs.writeFileSync(p,s,{flag:'wx'});
const once=(s,a,b)=>{if(s.split(a).length!==2)throw Error(`nonunique ${a}`);return s.replace(a,b)};
let state=read('internal/observationgate/subset_state.go').replaceAll('subsetState','ageState').replaceAll('subsetPending','agePending');
state=once(state,'subset *observationlearners.ConditionalForest, sequence','subset, age *observationlearners.ConditionalForest, sequence');
state=once(state,'observationlearners.RunConditionalObserver(subset, reader, reader.Epoch())',`observationlearners.RunConditionalObserver(chooseAgeObserver(subset, age, iw), reader, reader.Epoch())`);
state=once(state,'inside[1], inside[2], inside[3] = v, v, v',`inside[1], inside[2], inside[3] = v, v, v
        if age != nil {
            av, e := age.Forecast(p.Mask, p.Values)
            if e != nil { return observationpreserved.Prediction{}, e }
            inside[2], inside[3] = av, av
        }`);
state += `
// The two age slots are one algorithm with combined mass, not independent votes.
func chooseAgeObserver(retained, age *observationlearners.ConditionalForest, weights [4]float64) *observationlearners.ConditionalForest {
 if age != nil && weights[2]+weights[3] > weights[1] { return age }
 return retained
}
`;
write('internal/observationgate/age_state.go',state);
let journal=read('internal/observationgate/feedback_journal.go');
journal=once(journal,'const forecastJournalCapacity = 64\n','');
journal=once(journal,'type forecastJournalStats struct{ Applied, Stale, Censored, Pending int }\n','');
for(const [a,b]of [['forecastJournalEntry','ageJournalEntry'],['forecastJournal','ageJournal'],['subsetPending','agePending'],['subsetState','ageState']])journal=journal.replaceAll(new RegExp(`\\b${a}\\b`,'g'),b);
journal=once(journal,'enabled, split bool','enabled, split bool\n ageEnabled bool\n age *observationlearners.ConditionalForest');
journal=once(journal,'subset *observationlearners.ConditionalForest) error','subset, age *observationlearners.ConditionalForest) error');
journal=once(journal,'(subset != nil && models.Short == nil)','(subset != nil && models.Short == nil) || (age != nil && subset == nil)');
journal=once(journal,'g.models, g.subset = models, subset','g.models, g.subset, g.age = models, subset, age');
journal=once(journal,'p, e := s.predict(reader, g.models, g.subset, origin, seed)',`age := g.age
 if !g.ageEnabled { age = nil }
 p, e := s.predict(reader, g.models, g.subset, age, origin, seed)`);
write('internal/observationgate/age_journal.go',journal);
const src=read('internal/observationgate/delayed_learning_test.go');
let run=src.slice(src.indexOf('func delayedLearningRun('),src.indexOf('func TestDelayedLearningContracts('));
run=once(run,'func delayedLearningRun(','func ageLearningRun(');
run=once(run,') (delayedRecord, error)',') (ageRecord, error)');
run=once(run,'r := delayedRecord{Split: split, Scenario: name, Schedule: schedule, Index: index, StreamBase: seed, GateFirst: -1}','r := ageRecord{delayedRecord: delayedRecord{Split: split, Scenario: name, Schedule: schedule, Index: index, StreamBase: seed, GateFirst: -1}}');
run=once(run,'"journal_count"','"retained_control"');run=once(run,'"journal_retained_subset"','"age_challenger"');
run=once(run,'g := [2]forecastJournal{{base: base}, {base: base, enabled: true}}','g := [2]ageJournal{{base: base, enabled: true}, {base: base, enabled: true, ageEnabled: true}}');
run=once(run,'models.Version++',`age, support, e := fitAgeChallenger(step, audits)
                    if e != nil { return r, e }
                    if age != nil { r.AgeFits++; r.AgeSamples += support }
                    models.Version++`);
run=once(run,'g[i].publish(models, fit.model)','g[i].publish(models, fit.model, age)');
const imports=`// Mechanically derived v85 driver with the declared age challenger only.
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
write('internal/observationgate/age_learning_run_test.go',imports+run);
