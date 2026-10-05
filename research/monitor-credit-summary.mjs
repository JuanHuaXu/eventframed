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
  assert.equal(r.OriginalCost,p.MonitorCost+p.AuditCost);
  assert.equal(r.OriginalSplit,p.Arms[2].SplitAt);
  assert.equal(r.Masks.length,512);assert.equal(r.Values.length,512);assert.equal(r.Acquisition.length,512);
  let credit=0,old=0,spent=0,fullCount=0;
  const pop=x=>{let n=0;for(;x;x&=x-1)n++;return n;};
  for(let i=0;i<512;i++){
    const t=r.Acquisition[i],f=p.Frames[i],b=t.Bounded;
    assert.ok(Number.isInteger(b)&&b>=0&&b<=18);assert.equal(t.Audit,f.Audit);
    const full=f.Audit||credit>=18-b;assert.equal(t.Full,full);fullCount+=Number(full);
    const charge=full?18:b,allow=b+(f.Audit?18:0);
    credit+=allow-charge;old+=allow;spent+=charge;
    assert.equal(t.Charge,charge);assert.equal(t.Allowance,allow);assert.equal(t.Credit,credit);assert.ok(credit>=0&&spent<=old);
    assert.equal(r.Masks[i].reduce((n,m)=>n+pop(m),0),charge);
    if(full)assert.deepEqual(r.Masks[i],[511,511]);
    assert.deepEqual(r.Values[i],[f.X&r.Masks[i][0],f.RX&r.Masks[i][1]]);
  }
  assert.equal(old,r.OriginalCost);assert.equal(spent,r.CandidateCost);assert.equal(fullCount,r.SelectedPairs);
  const eligible=p.Released,a=r.Candidate;
  assert.equal(r.ArrivedPairs,eligible.length);assert.deepEqual(a.Trace.map(t=>t.Origin),eligible);
  for(const t of a.Trace){const f=p.Frames[t.Origin];assert.ok(!f.Missing&&f.Arrival<=t.Clock);if(!r.Acquisition[t.Origin].Full){assert.equal(t.LiveCorrect,(f.Baseline>=.5)===f.Y);assert.equal(t.ReferenceCorrect,(f.Reference>=.5)===f.RY);}}
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
const result={rawSHA256:hash(raw),parentSHA256:d.ParentSHA256,sources:Object.keys(d.Sources).length,screens,cells,limits:'Consumed shadow shared-read credit ledger. No closed-loop gains, fresh confirmation, periodic-stream robustness, or universal cost guarantee. Means of split clocks exclude misses.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({hash:result.rawSHA256,sources:result.sources,screens,delayed}));
