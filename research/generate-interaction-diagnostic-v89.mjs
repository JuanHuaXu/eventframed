import fs from 'node:fs';
let s=fs.readFileSync('internal/observationgate/age_breadth_run_test.go','utf8');
function once(a,b){if(s.split(a).length!==2)throw Error(`nonunique ${a}`);s=s.replace(a,b)}
once('func ageBreadthRun(','func interactionDiagnosticRun(');
once('cfg subsetBreadthCase) (ageRecord, error)','cfg subsetBreadthCase, diagnostic *interactionDiagnostic) (ageRecord, error)');
once('\t\tvar a, b observation.Sample',`        if e := diagnostic.prepare(step, x, cfg, g[1].entries[step%forecastJournalCapacity].forecast, models, base, g[1].subset, g[1].age, g[1].mix, g[1].inner); e != nil { return r, e }
        var a, b observation.Sample`);
once('\t\tfor i, p := range predictions {',`        if e := diagnostic.observe(step, y); e != nil { return r, e }
        for i, p := range predictions {`);
fs.writeFileSync('internal/observationgate/interaction_diagnostic_run_test.go',s,{flag:'wx'});
