import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-allowance-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,6);
const seen=new Set();
const refusal='insufficient research computation allowance';
for(const a of d.Arms){
 const key=[a.Trial,a.Arm].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.Traces.length,a.Started);assert.equal(new Set(a.Traces.map(t=>t.ID)).size,a.Started);
 for(const t of a.Traces){assert(Math.abs(t.EntryRemainingMS-t.ComputeMS-t.ExitRemainingMS)<1e-7);assert(t.ComputeMS>=0);if(a.Arm==='allowance')assert.equal(t.FitError===refusal,t.EntryRemainingMS<65);else assert.notEqual(t.FitError,refusal);}
 assert.equal(a.Checked,a.Traces.filter(t=>!t.FitError).length);
 const s=a.Status;assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);assert.equal(s.Depth,0);assert(!s.Running);
}
const summary=a=>({completed:a.Status.Completed,p99:a.P99/1e6,interrupted:a.Interrupted,refused:a.Traces.filter(t=>t.FitError===refusal).length,processorMS:a.Traces.reduce((v,t)=>v+t.ComputeMS,0),dropped:a.Status.Dropped,unexpected:a.Traces.filter(t=>t.FitError&&!['context deadline exceeded',refusal].includes(t.FitError)).length});
const rows=[];
for(let trial=0;trial<3;trial++){
 const b=d.Arms.find(a=>a.Trial===trial&&a.Arm==='batch'),a=d.Arms.find(a=>a.Trial===trial&&a.Arm==='allowance');assert(a&&b);
 const base=summary(b),candidate=summary(a);
 rows.push({trial,base,candidate,passed:!(a.Errors?.length||b.Errors?.length)&&a.Overlap>0&&b.Overlap>0&&candidate.unexpected===0&&candidate.completed>=52&&candidate.p99<=1.1*base.p99});
}
console.log(JSON.stringify({rows,passed:rows.every(r=>r.passed)},null,2));
