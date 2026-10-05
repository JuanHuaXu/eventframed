import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import path from 'node:path';
const root=path.resolve(import.meta.dirname,'..');
const data=JSON.parse(fs.readFileSync(path.join(import.meta.dirname,'segment-transform-burst-results.json')));
for(const [file,hash] of Object.entries(data.sources)) {
 const bytes=fs.readFileSync(path.resolve(root,'internal/observationlearners',file));
 assert.equal(crypto.createHash('sha256').update(bytes).digest('hex'),hash);
}
assert.equal(data.bursts.length,18);
const seen=new Set();
const rows=data.bursts.map(b=>{
 const key=[b.trial,b.workers,b.transform].join(':');assert(!seen.has(key));seen.add(key);
 assert([0,1,2].includes(b.trial)&&[1,2,4].includes(b.workers)&&typeof b.transform==='boolean');
 assert.equal(b.jobs.length,12);
 b.jobs.forEach((j,id)=>{
  assert.equal(j.id,id);
  for(const v of [j.queue_ms,j.compute_ms,j.completion_ms])assert(Number.isFinite(v)&&v>=0);
  assert(Math.abs(j.queue_ms+j.compute_ms-j.completion_ms)<1e-8);
 });
 return {trial:b.trial,workers:b.workers,transform:b.transform,makespan:Math.max(...b.jobs.map(j=>j.completion_ms)),by100:b.jobs.filter(j=>j.completion_ms<=100).length,by500:b.jobs.filter(j=>j.completion_ms<=500).length};
});
const summaries=[];
for(const workers of [1,2,4]) {
 const arm=fast=>rows.filter(r=>r.workers===workers&&r.transform===fast);
 const direct=arm(false),fast=arm(true);
 const median=a=>a.slice().sort((x,y)=>x-y)[1];
 const gain=1-median(fast.map(r=>r.makespan))/median(direct.map(r=>r.makespan));
 const nonharm=direct.every(a=>{const b=fast.find(b=>b.trial===a.trial);return b.by100>=a.by100&&b.by500>=a.by500;});
 summaries.push({workers,medianMakespanGain:gain,deadlineNonharm:nonharm,passed:gain>=.1&&nonharm});
}
console.log(JSON.stringify({scope:'isolated fitter burst, not service latency',rows,summaries,passed:summaries.every(s=>s.passed)},null,2));
