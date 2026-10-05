import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';

function input(file) {const stream=fs.createReadStream(file),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {iterator:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const raw=input(process.argv[2]),source=input(process.argv[3]),baseline=input(process.argv[4]);
const header=JSON.parse((await raw.iterator.next()).value),sourceHeader=JSON.parse((await source.iterator.next()).value),baselineHeader=JSON.parse((await baseline.iterator.next()).value);
assert.equal(header.Version,'regime-query-disjoint-v1');assert.equal(sourceHeader.Version,'soft-learners-v120');assert.equal(baselineHeader.Version,'regime-query-outcome-v1');
for(const [k,v] of Object.entries({Records:2688,Workers:4,ProbeStart:137,ProbeCount:8}))assert.equal(header[k],v);
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for(const [file,expected] of Object.entries(header.Hashes)){const absolute=path.resolve('internal/observationlearners',file);assert(absolute.startsWith(process.cwd()+path.sep));assert.equal(sha(absolute),expected);}
const records=[],ids=new Set();let checks=0,oldFits=0,newFits=0,extraBatches=0;
for(;;){
  const a=await raw.iterator.next(),b=await source.iterator.next(),c=await baseline.iterator.next();
  if(a.done||b.done||c.done){assert(a.done&&b.done&&c.done);break;}
  const r=JSON.parse(a.value),s=JSON.parse(b.value),old=JSON.parse(c.value);assert(!r.Error);assert.deepEqual(r.Original,old);
  const id=`${s.Phase}:${s.Case}:${s.Index}:${s.Schedule}`;assert(!ids.has(id));ids.add(id);
  const original=auditRegimeOutcome(r.Original,s,153),candidate=auditRegimeOutcome(r.Candidate,s,137);
  assert.deepEqual(original.costs,candidate.costs);assert.deepEqual(r.Original.Decision.Pool,r.Candidate.Decision.Pool);assert.deepEqual(r.Original.Decision.Origins,r.Candidate.Decision.Origins);
  const ov=r.Original.Decision.Values??[],cv=r.Candidate.Decision.Values??[];assert.equal(ov.length,cv.length);
  for(let i=0;i<ov.length;i++)for(let y=0;y<2;y++){assert(Math.abs(ov[i].Mass[y]-cv[i].Mass[y])<1e-10);assert(Math.abs(ov[i].LogEvidence[y]-cv[i].LogEvidence[y])<1e-10);}
  for(let arm=0;arm<3;arm++){
    assert.equal(candidate.selected[arm],original.selected[arm]);assert.equal(candidate.brier[arm],original.brier[arm]);
    for(const field of ['Origins','AtPublication','Predictions','LogEvidence','Redundant'])assert.deepEqual(r.Candidate[field][arm],r.Original[field][arm]);
  }
  const expectedExtra=original.poolSize>0?1:0;assert.equal(r.ExtraBatches,expectedExtra);extraBatches+=r.ExtraBatches;
  oldFits+=r.Original.ActualFits;newFits+=r.Candidate.ActualFits;
  const selected=candidate.selected[3];
  const exact=(probes,x)=>probes.filter(p=>p===x).length/probes.length;
  const overlap=selected<0?null:{
    virtual:exact(r.Candidate.Decision.Probes,s.Steps[selected].X),
    future:exact(s.Steps.slice(161,192).map(v=>v.X),s.Steps[selected].X),
    originInProbes:Number(selected>=137&&selected<=144)};
  records.push({phase:s.Phase,case:s.Case,index:s.Index,schedule:s.Schedule,
    brier:[...original.brier,candidate.brier[3]],costs:[...original.costs,candidate.costs[3]],
    redundant:[...original.redundant,candidate.redundant[3]],supportChanged:[...original.supportChanged,candidate.supportChanged[3]],
    selectionChanged:Number(original.selected[3]!==selected),selected:[...original.selected,selected],
    originalPredictedGain:original.predictedGain[3],candidatePredictedGain:candidate.predictedGain[3],overlap});
  checks+=8*31;
}
assert.equal(records.length,2688);assert.equal(checks,666624);
const inputSHA256=source.hash.digest('hex'),baselineSHA256=baseline.hash.digest('hex');
assert.equal(inputSHA256,header.InputSHA256);assert.equal(inputSHA256,baselineHeader.InputSHA256);assert.equal(inputSHA256,'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');assert.equal(baselineSHA256,header.BaselineSHA256);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  groups.push({phase,case:c,schedule,brier:[0,1,2,3,4].map(a=>mean(rs.map(r=>r.brier[a]))),
    candidateGains:[0,1,2,3].map(a=>ci(rs.map(r=>r.brier[a]-r.brier[4]))),
    costs:[0,1,2,3,4].map(a=>rs.reduce((s,r)=>s+r.costs[a],0)),redundant:[0,1,2,3,4].map(a=>rs.reduce((s,r)=>s+r.redundant[a],0)),
    selectionChanged:rs.reduce((s,r)=>s+r.selectionChanged,0)});
}
const delayed=records.filter(r=>r.schedule===1);assert.equal(delayed.length,1344);
const totals={extraBatches,originalPublicationFits:oldFits,candidatePublicationFits:newFits,
  selectionChanged:delayed.reduce((s,r)=>s+r.selectionChanged,0),
  costs:[0,1,2,3,4].map(a=>records.reduce((s,r)=>s+r.costs[a],0)),redundant:[0,1,2,3,4].map(a=>records.reduce((s,r)=>s+r.redundant[a],0)),
  delayedMeanBrier:[0,1,2,3,4].map(a=>mean(delayed.map(r=>r.brier[a]))),
  overlap:Object.fromEntries(['virtual','future','originInProbes'].map(k=>[k,mean(delayed.map(r=>r.overlap[k]))]))};
console.log(JSON.stringify({scope:'Controlled one-decision probe comparison on consumed data; not a full-stream or whole-goal pass/fail',inputSHA256,baselineSHA256,artifactSHA256:raw.hash.digest('hex'),
  scriptSHA256:sha(import.meta.filename),auditSHA256:sha(new URL('./regime-query-outcome-audit.mjs',import.meta.url)),checks,totals,records,groups}));
