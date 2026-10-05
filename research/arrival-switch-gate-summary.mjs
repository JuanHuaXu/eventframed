import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input, output]=process.argv.slice(2);
const raw=fs.readFileSync(input),d=JSON.parse(raw);
assert.equal(d.Records.length,256);
for(const r of d.Records){
  assert.ok(r.AccuracyBefore>=0&&r.AccuracyBefore<=1&&r.AccuracyAfter>=0&&r.AccuracyAfter<=1);
  assert.ok(Math.abs(r.AccuracyBefore-r.AccuracyAfter)<=r.ConditionalTV+1e-12);
  if(r.Case.startsWith('stable_')) assert.equal(r.ConditionalTV,0);
  let prev=-1, clock=-1;
  for(const t of r.Trace){assert.ok(t.Origin>prev&&t.Clock>=clock);prev=t.Origin;clock=t.Clock;assert.ok(Number.isFinite(t.LogWealth));}
  assert.equal(r.FirstAllow,r.Trace.find(t=>t.Allow)?.Clock??-1);
  assert.equal(r.FirstNomination,r.Trace.find(t=>t.Nominate)?.Clock??-1);
  assert.equal(r.SplitAt,r.Trace.find(t=>t.Allow&&t.Nominate)?.Clock??-1);
}
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(d.Records.map(r=>r.Case))])for(const schedule of ['Immediate','Delayed']){
  const a=d.Records.filter(r=>r.Phase===phase&&r.Case===name&&r.Schedule===schedule);assert.equal(a.length,16);
  cells.push({phase,case:name,schedule,n:a.length,meanReferenceAccuracy:mean(a.map(r=>r.AccuracyBefore)),meanLiveAccuracy:mean(a.map(r=>r.AccuracyAfter)),meanConditionalTV:mean(a.map(r=>r.ConditionalTV)),exactlyInvisible:a.filter(r=>Math.abs(r.AccuracyBefore-r.AccuracyAfter)<1e-12).length,allowed:a.filter(r=>r.FirstAllow>=0).length,nominated:a.filter(r=>r.FirstNomination>=0).length,split:a.filter(r=>r.SplitAt>=0).length,maxAbsAccuracyGap:Math.max(...a.map(r=>Math.abs(r.AccuracyBefore-r.AccuracyAfter)))});
}
const result={diagnosticSHA256:crypto.createHash('sha256').update(raw).digest('hex'),parentSHA256:d.RawSHA256,cells,limits:'Exhaustive finite generator expectations and consumed gate trace diagnosis, not power estimates for deployment or an empirical null certificate.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.schedule==='Delayed')));
