import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./segment-paced-fit-results.json',import.meta.url)));
for(const [f,h] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../internal/observationlearners',f))).digest('hex'),h);
assert.equal(d.Arms.length,12);const seen=new Set();
for(const a of d.Arms){
 const key=[a.Rate,a.Trial,a.Arm].join(':');assert(!seen.has(key));seen.add(key);
 assert.equal(a.ReadNS.length,64);assert.equal(a.ScheduledReadNS.length,64);
 if(a.Arm==='off'){assert.equal(a.Traces,null);assert.equal(a.Started,0);assert.equal(a.Checked,0);}
 else {assert.equal(a.Traces.length,a.Started);assert.equal(a.Checked,a.Traces.filter(t=>!t.FitError).length);}
 if(a.Arm!=='off'){const s=a.Status;assert.equal(s.Accepted+s.Dropped,64);assert.equal(s.Completed+s.Stale+s.Failed+s.Cancelled,s.Accepted);assert.equal(s.Depth,0);assert(!s.Running);}
}
const rows=[];
for(const rate of [10,40])for(let trial=0;trial<3;trial++){
 const on=d.Arms.find(a=>a.Rate===rate&&a.Trial===trial&&a.Arm==='table'),off=d.Arms.find(a=>a.Rate===rate&&a.Trial===trial&&a.Arm==='off');assert(on&&off);
 const late=on.Traces.filter(t=>!t.FitError&&t.FromOfferMS>100).length;
 const onTimeLower=Math.max(0,on.Status.Completed-late);
 const onP99=Math.max(...on.ScheduledReadNS)/1e6,offP99=Math.max(...off.ScheduledReadNS)/1e6;
 rows.push({rate,trial,completed:on.Status.Completed,drops:on.Status.Dropped,stale:on.Status.Stale,lateReturns:late,onTimeLower,onP99,offP99,passed:!(on.Errors?.length||off.Errors?.length)&&on.Overlap>0&&off.Overlap>0&&on.Status.Completed>=52&&onTimeLower>=52&&onP99<=1.1*offP99});
}
console.log(JSON.stringify({rows,passed:rows.every(r=>r.passed),boundary:'on-time at processor return; final validation time untraced'},null,2));
