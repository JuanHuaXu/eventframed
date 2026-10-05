import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-transform-service-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,6);
const seen=new Set();
for(const a of d.Arms){
 const key=[a.Trial,a.Enabled].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.ReadNS.length,64);assert.equal(a.WriteNS.length,16);
 assert.equal(Math.max(...a.ReadNS),a.P99);
 assert(a.ReadNS.concat(a.WriteNS).every(x=>Number.isFinite(x)&&x>=0));
 if(a.Enabled){const s=a.Status;assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);assert.equal(s.Depth,0);assert.equal(s.Running,false);}
}
const rows=[];
for(let trial=0;trial<3;trial++){
 const off=d.Arms.find(a=>a.Trial===trial&&!a.Enabled),on=d.Arms.find(a=>a.Trial===trial&&a.Enabled);
 assert(off&&on);
 const integrity=!(off.Errors?.length||on.Errors?.length)&&off.Overlap>0&&on.Overlap>0&&on.Status.Failed===0;
 const ratio=on.P99/off.P99;
 rows.push({trial,integrity,p99Ratio:ratio,latencyPass:ratio<=1.1,completed:on.Status.Completed,completionPass:on.Status.Completed>=52,passed:integrity&&ratio<=1.1&&on.Status.Completed>=52});
}
console.log(JSON.stringify({rows,passed:rows.every(r=>r.passed),scope:'actual scheduler with fixture fitter; no task usefulness validation'},null,2));
