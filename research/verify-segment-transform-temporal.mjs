import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-transform-temporal-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,6);
const seen=new Set();
for(const a of d.Arms){
 const key=[a.Trial,a.Temporal].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.ReadNS.length,64);assert.equal(a.WriteNS.length,16);assert.equal(Math.max(...a.ReadNS),a.P99);
 const s=a.Status;
 assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);
 assert.equal(s.Depth,0);assert.equal(s.Running,false);
 assert(a.Interrupted<=a.Started&&a.Started<=s.Accepted);
}
const rows=[];
for(let trial=0;trial<3;trial++){
 const strict=d.Arms.find(a=>a.Trial===trial&&!a.Temporal),temporal=d.Arms.find(a=>a.Trial===trial&&a.Temporal);assert(strict&&temporal);
 const ratio=temporal.P99/strict.P99;
 const integrity=!(strict.Errors?.length||temporal.Errors?.length)&&strict.Overlap>0&&temporal.Overlap>0;
 rows.push({trial,integrity,p99Ratio:ratio,latencyPass:ratio<=1.1,completed:temporal.Status.Completed,completionPass:temporal.Status.Completed>=52&&temporal.Status.Completed>=strict.Status.Completed,started:temporal.Started,interrupted:temporal.Interrupted,dropped:temporal.Status.Dropped});
}
console.log(JSON.stringify({rows,passed:rows.every(r=>r.integrity&&r.latencyPass&&r.completionPass)},null,2));
