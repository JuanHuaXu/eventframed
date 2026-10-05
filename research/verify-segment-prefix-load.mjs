import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-prefix-load-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,12);const seen=new Set();
for(const a of d.Arms){
 const key=[a.Distinct,a.Trial,a.Arm].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.Traces.length,a.Started);assert.equal(new Set(a.Traces.map(t=>t.ID)).size,a.Started);
 assert.equal(a.Checked,a.Traces.filter(t=>!t.FitError).length);
 for(const t of a.Traces)assert(Math.abs(t.EntryRemainingMS-t.ComputeMS-t.ExitRemainingMS)<1e-7);
 const s=a.Status;assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);assert.equal(s.Depth,0);assert(!s.Running);
 if(a.Distinct)assert.equal(a.Hits,0);
}
const rows=[];
for(const invalidate of [false,true])for(let trial=0;trial<3;trial++){
 const a=d.Arms.find(a=>a.Distinct===invalidate&&a.Trial===trial&&a.Arm==='prefix'),b=d.Arms.find(a=>a.Distinct===invalidate&&a.Trial===trial&&a.Arm==='table');assert(a&&b);
 rows.push({invalidate,trial,controlCompleted:b.Status.Completed,completed:a.Status.Completed,preparations:a.Fits,reuseAttempts:a.Hits,verifiedReturns:a.Checked,drops:a.Status.Dropped,controlP99:b.P99/1e6,prefixP99:a.P99/1e6,passed:!(a.Errors?.length||b.Errors?.length)&&a.Overlap>0&&b.Overlap>0&&a.Status.Completed>=52&&a.Status.Completed>=b.Status.Completed&&a.P99<=1.1*b.P99});
}
console.log(JSON.stringify({rows,passed:rows.every(r=>r.passed)},null,2));
