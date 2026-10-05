import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const [streamPath,marginPath,output]=process.argv.slice(2);
assert.ok(streamPath&&marginPath&&output);
const raw=fs.readFileSync(streamPath),marginRaw=fs.readFileSync(marginPath);
assert.equal(hash(raw),'c9a21a020a2321ca46481d0c145205cd8e3f78349cd4475494f87d72fd337451');
const stream=JSON.parse(raw),single=JSON.parse(marginRaw);
assert.equal(stream.records.length,256);
assert.equal(single.records.length,256);
const starts=[1,2,4,8,16,32,64,128];
const rate=.5,threshold=20,diameter=.15;

function firstCrossing(scores){
  const products=Array(starts.length).fill(1);
  for(let i=0;i<scores.length;i++){
    const row=scores[i],w=row.referenceCorrect-row.liveCorrect;
    assert.ok(Number.isInteger(w)&&w>=-1&&w<=1);
    const factor=1+rate*(w-diameter);
    assert.ok(factor>=0);
    for(let j=0;j<starts.length;j++)if(i+1>=starts[j])products[j]*=factor;
    const wealth=products.reduce((a,b)=>a+b,0)/starts.length;
    if(wealth>=threshold)return {clock:row.release,origin:row.origin,auditUpdates:i+1,wealth};
  }
  return null;
}
assert.equal(firstCrossing([]),null);
assert.equal(firstCrossing(Array.from({length:32},(_,i)=>({origin:i,release:i,referenceCorrect:0,liveCorrect:1}))),null);
assert.ok(firstCrossing(Array.from({length:64},(_,i)=>({origin:i,release:i,referenceCorrect:1,liveCorrect:0}))));

const records=[];
for(let k=0;k<stream.records.length;k++){
  const row=stream.records[k],prior=single.records[k];
  for(const key of ['phase','case','index','schedule'])assert.equal(row[key],prior[key]);
  const nomination=firstCrossing(row.scores);
  const changed=row.case.includes('_to_');
  records.push({phase:row.phase,case:row.case,index:row.index,schedule:row.schedule,
    nomination,singleMargin:prior.margin015,externalSplitClock:row.split,
    changed,beforeTrueChange:changed&&nomination!==null&&nomination.clock<256});
}
const mean=a=>a.length?a.reduce((s,x)=>s+x,0)/a.length:null;
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of ['stable_majority3','stable_parity4','majority_to_parity','parity_to_majority']){
  for(const schedule of ['Immediate','Delayed']){
    const rows=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);
    assert.equal(rows.length,16);
    cells.push({phase,case:name,schedule,
      nominated:rows.filter(r=>r.nomination).length,
      by511:rows.filter(r=>r.nomination&&r.nomination.clock<=511).length,
      by543:rows.filter(r=>r.nomination&&r.nomination.clock<=543).length,
      beforeTrueChange:rows.filter(r=>r.beforeTrueChange).length,
      singleMarginBy511:rows.filter(r=>r.singleMargin&&r.singleMargin.clock<=511).length,
      externalSplits:rows.filter(r=>r.externalSplitClock>=0).length,
      meanClockAmongDetected:mean(rows.flatMap(r=>r.nomination?[r.nomination.clock]:[])),
      meanAuditUpdatesAmongDetected:mean(rows.flatMap(r=>r.nomination?[r.nomination.auditUpdates]:[]))});
  }
}
const result={streamSHA256:hash(raw),marginSHA256:hash(marginRaw),scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),starts,rate,diameter,threshold,cells,records,
  limits:'Consumed synthetic audit nomination only. Fixed mixture is a scalar e-process under its declared conditional null; it does not establish target-law diameter, Anti-Pigeon authorization, onset coverage, or predictive benefit.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells));
