import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,output]=process.argv.slice(2);
const raw=fs.readFileSync(input),d=JSON.parse(raw),praw=fs.readFileSync(parent);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(hash(praw),d.ParentSHA256);assert.equal(d.Consumed,true);
for(const [p,s] of Object.entries(d.Sources)){assert.equal(hash(s),d.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),d.Hashes[p]);}
const [, ...parents]=praw.toString().trim().split('\n').map(JSON.parse);
const key=r=>[r.Phase,r.Case,r.Index].join('/');
const map=new Map(parents.map(r=>[key(r),r]));
assert.equal(d.Records.length,256);
for(const r of d.Records){
  const s=map.get(key(r))[r.Schedule];
  assert.equal(r.AuditPairs,s.Frames.filter(f=>f.Audit).length);
  const eligible=s.Released.filter(i=>s.Frames[i].Audit);
  assert.equal(r.ArrivedAuditPairs,eligible.length);
  for(let a=0;a<4;a++){
    const z=r.Arms[a];assert.deepEqual(z.Trace.map(t=>t.Origin),(a===1||a===2)?eligible:s.Released);
    assert.equal(z.FirstAllow,z.Trace.find(t=>t.Allow)?.Clock??-1);
    assert.equal(z.FirstNomination,z.Trace.find(t=>t.Nominate)?.Clock??-1);
    assert.equal(z.SplitAt,z.Trace.find(t=>t.Allow&&t.Nominate)?.Clock??-1);
    for(const t of z.Trace){const f=s.Frames[t.Origin];assert.ok(!f.Missing&&f.Arrival<=t.Clock);if(a<2){assert.equal(t.ReferenceCorrect,(f.Reference>=.5)===f.RY);assert.equal(t.LiveCorrect,(f.Baseline>=.5)===f.Y);}}
  }
  assert.equal(r.Arms[0].SplitAt,s.Arms[2].SplitAt);
}
const mean=a=>a.length?a.reduce((s,v)=>s+v,0)/a.length:null,cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(d.Records.map(r=>r.Case))])for(const schedule of ['Immediate','Delayed']){
  const rs=d.Records.filter(r=>r.Phase===phase&&r.Case===name&&r.Schedule===schedule);assert.equal(rs.length,16);
  cells.push({phase,case:name,schedule,n:rs.length,accuracyBefore:[0,1].map(a=>mean(rs.map(r=>r.AccuracyBefore[a]))),accuracyAfter:[0,1].map(a=>mean(rs.map(r=>r.AccuracyAfter[a]))),meanAuditPairs:mean(rs.map(r=>r.AuditPairs)),meanArrivedAuditPairs:mean(rs.map(r=>r.ArrivedAuditPairs)),arms:[0,1,2,3].map(a=>({allowed:rs.filter(r=>r.Arms[a].FirstAllow>=0).length,nominated:rs.filter(r=>r.Arms[a].FirstNomination>=0).length,split:rs.filter(r=>r.Arms[a].SplitAt>=0).length,splitBy511:rs.filter(r=>r.Arms[a].SplitAt>=0&&r.Arms[a].SplitAt<=511).length,meanDetectedClock:mean(rs.filter(r=>r.Arms[a].SplitAt>=0).map(r=>r.Arms[a].SplitAt))}))});
}
const delayed=cells.filter(c=>c.schedule==='Delayed');
const screens={reverseImproved:delayed.filter(c=>c.case==='parity_to_majority').every(c=>c.arms[3].split>c.arms[0].split),forwardRetained:delayed.filter(c=>c.case==='majority_to_parity').every(c=>c.arms[3].split>=c.arms[0].split),noAddedStableSplits:cells.filter(c=>c.case.startsWith('stable_')).every(c=>c.arms[3].split<=c.arms[0].split)};
const result={rawSHA256:hash(raw),parentSHA256:d.ParentSHA256,sources:Object.keys(d.Sources).length,screens,cells,limits:'Consumed shadow gate comparison, not closed-loop recovery, fresh confirmation, or equal-cost observation-policy evidence. Detected-clock means exclude misses.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({hash:result.rawSHA256,sources:result.sources,screens,delayed}));
