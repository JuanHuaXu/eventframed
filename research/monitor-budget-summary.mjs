import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,output]=process.argv.slice(2);
const raw=fs.readFileSync(input),d=JSON.parse(raw),praw=fs.readFileSync(parent);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(d.ParentSHA256,hash(praw));assert.equal(d.Consumed,true);
for(const [p,s]of Object.entries(d.Sources)){assert.equal(hash(s),d.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),d.Hashes[p]);}
const [, ...parents]=praw.toString().trim().split('\n').map(JSON.parse);
const key=r=>[r.Phase,r.Case,r.Index].join('/'),map=new Map(parents.map(r=>[key(r),r]));
assert.equal(d.Records.length,256);
for(const r of d.Records){
  const p=map.get(key(r))[r.Schedule];
  const selected=p.Frames.map((f,i)=>i%2===0||f.Audit);
  assert.equal(r.SelectedPairs,selected.filter(Boolean).length);
  assert.equal(r.CandidateCost,r.SelectedPairs*18);
  assert.equal(r.OriginalCost,p.MonitorCost+p.AuditCost);
  assert.equal(r.OriginalSplit,p.Arms[2].SplitAt);
  assert.equal(r.Masks.length,512);
  selected.forEach((s,i)=>assert.deepEqual(r.Masks[i],s?[511,511]:[0,0]));
  const eligible=p.Released.filter(i=>selected[i]),a=r.Candidate;
  assert.equal(r.ArrivedPairs,eligible.length);assert.deepEqual(a.Trace.map(t=>t.Origin),eligible);
  for(const t of a.Trace){const f=p.Frames[t.Origin];assert.ok(!f.Missing&&f.Arrival<=t.Clock);}
  assert.equal(a.FirstAllow,a.Trace.find(t=>t.Allow)?.Clock??-1);
  assert.equal(a.FirstNomination,a.Trace.find(t=>t.Nominate)?.Clock??-1);
  assert.equal(a.SplitAt,a.Trace.find(t=>t.Allow&&t.Nominate)?.Clock??-1);
}
const mean=a=>a.length?a.reduce((s,v)=>s+v,0)/a.length:null,cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(d.Records.map(r=>r.Case))])for(const schedule of ['Immediate','Delayed']){
  const rs=d.Records.filter(r=>r.Phase===phase&&r.Case===name&&r.Schedule===schedule);assert.equal(rs.length,16);
  cells.push({phase,case:name,schedule,n:16,originalSplits:rs.filter(r=>r.OriginalSplit>=0).length,candidateSplits:rs.filter(r=>r.Candidate.SplitAt>=0).length,candidateBy511:rs.filter(r=>r.Candidate.SplitAt>=0&&r.Candidate.SplitAt<=511).length,meanDetectedClock:mean(rs.filter(r=>r.Candidate.SplitAt>=0).map(r=>r.Candidate.SplitAt)),originalCoordinates:mean(rs.map(r=>r.OriginalCost/512)),candidateCoordinates:mean(rs.map(r=>r.CandidateCost/512)),maxCostRatio:Math.max(...rs.map(r=>r.CandidateCost/r.OriginalCost)),arrivedPairs:mean(rs.map(r=>r.ArrivedPairs))});
}
const delayed=cells.filter(c=>c.schedule==='Delayed');
const screens={reverseImproved:delayed.filter(c=>c.case==='parity_to_majority').every(c=>c.candidateSplits>c.originalSplits),forwardRetained:delayed.filter(c=>c.case==='majority_to_parity').every(c=>c.candidateSplits>=c.originalSplits),noAddedStableSplits:cells.filter(c=>c.case.startsWith('stable_')).every(c=>c.candidateSplits<=c.originalSplits),perTrajectoryCost:d.Records.every(r=>r.CandidateCost<=r.OriginalCost)};
const result={rawSHA256:hash(raw),parentSHA256:d.ParentSHA256,sources:Object.keys(d.Sources).length,screens,cells,limits:'Consumed shadow budget reallocation. No closed-loop gains, fresh confirmation, periodic-stream robustness, or universal cost guarantee. Means of split clocks exclude misses.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({hash:result.rawSHA256,sources:result.sources,screens,delayed}));
