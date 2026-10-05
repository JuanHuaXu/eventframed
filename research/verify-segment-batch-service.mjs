import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-batch-service-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,18);
const seen=new Set();
for(const a of d.Arms){
 const key=[a.Distinct,a.Trial,a.Arm].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.ReadNS.length,64);assert.equal(a.WriteNS.length,16);assert.equal(a.P99,Math.max(...a.ReadNS));
 const s=a.Status;
 if(a.Arm!=='off'){assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);assert.equal(s.Depth,0);assert(!s.Running);assert(s.Completed<=a.Checked);}
 if(a.Distinct)assert.equal(a.Hits,0);
 if(a.Arm==='batch')assert.equal(a.Checked,a.Hits+a.Fits);
}
const rows=[];
for(const distinct of [false,true])for(let trial=0;trial<3;trial++){
 const get=arm=>d.Arms.find(a=>a.Distinct===distinct&&a.Trial===trial&&a.Arm===arm);
 const off=get('off'),direct=get('direct'),memo=get('batch');assert(off&&direct&&memo);
 const integrity=[off,direct,memo].every(a=>!a.Errors?.length&&a.Overlap>0);
 const row={distinct,trial,integrity,offMS:off.P99/1e6,directMS:direct.P99/1e6,batchMS:memo.P99/1e6,completed:memo.Status.Completed,directCompleted:direct.Status.Completed,hits:memo.Hits,fits:memo.Fits};
 row.passed=integrity&&row.completed>=52&&row.completed>=row.directCompleted&&memo.P99<=1.1*off.P99&&memo.P99<=1.1*direct.P99;
 rows.push(row);
}
console.log(JSON.stringify({rows,passed:rows.every(r=>r.passed)},null,2));
