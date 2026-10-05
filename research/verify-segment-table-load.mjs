import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-table-load-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,6);
const seen=new Set();const rows=[];
for(const a of d.Arms){
 const key=[a.Trial,a.Arm].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.Traces.length,a.Started);assert.equal(new Set(a.Traces.map(t=>t.ID)).size,a.Started);
 for(const t of a.Traces){assert(t.ID>=0&&t.ID<64);assert(Math.abs(t.EntryRemainingMS-t.ComputeMS-t.ExitRemainingMS)<1e-7);assert(t.ComputeMS>=0);}
 const success=a.Traces.filter(t=>!t.FitError),interrupted=a.Traces.filter(t=>t.ContextError);
 assert.equal(success.length,a.Checked);assert.equal(interrupted.length,a.Interrupted);
 const s=a.Status;assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);assert.equal(s.Depth,0);assert(!s.Running);
 rows.push({trial:a.Trial,arm:a.Arm,started:a.Started,notStarted:s.Accepted-a.Started,completed:s.Completed,dropped:s.Dropped,under10:a.Traces.filter(t=>t.EntryRemainingMS<10).length,under25:a.Traces.filter(t=>t.EntryRemainingMS<25).length,under40:a.Traces.filter(t=>t.EntryRemainingMS<40).length,interrupted:a.Interrupted,entryRange:[Math.min(...a.Traces.map(t=>t.EntryRemainingMS)),Math.max(...a.Traces.map(t=>t.EntryRemainingMS))],successfulCompute:success.map(t=>t.ComputeMS),postReturnRejections:success.length-s.Completed});
}
const verdict=[];
for(let trial=0;trial<3;trial++){
 const a=d.Arms.find(a=>a.Trial===trial&&a.Arm==='table'),b=d.Arms.find(a=>a.Trial===trial&&a.Arm==='batch');assert(a&&b);
 verdict.push({trial,ratio:a.P99/b.P99,passed:!(a.Errors?.length||b.Errors?.length)&&a.Overlap>0&&b.Overlap>0&&a.Status.Completed>=52&&a.Status.Completed>=b.Status.Completed&&a.P99<=1.1*b.P99});
}
console.log(JSON.stringify({rows,verdict,passed:verdict.every(v=>v.passed)},null,2));

