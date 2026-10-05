import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';
import {fitQueryShrinkage,shrinkQueryMass,queryMixtureScore,chooseShrinkage} from './query-shrinkage.mjs';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';
const [rawPath,sourcePath,popPath,samplePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const pop=JSON.parse(fs.readFileSync(popPath)),sample=JSON.parse(fs.readFileSync(samplePath));assert.equal(pop.records.length,2688);assert.equal(sample.records.length,2688);
function input(path){const stream=fs.createReadStream(path),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[rawPath,sourcePath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
assert.equal(headers[0].Version,'regime-query-disjoint-v1');assert.equal(headers[1].Version,'soft-learners-v120');
const runs=[],near=(a,b)=>assert(Math.abs(a-b)<1e-12),mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(s=>s.done)){assert(next.every(s=>s.done));break;}
  const [raw,source]=next.map(s=>JSON.parse(s.value)),i=runs.length,p=pop.records[i],t=sample.records[i];assert(p&&!raw.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(source[upper],p[lower]);assert.equal(p[lower],t[lower]);}
  auditRegimeOutcome(raw.Original,source);const d=raw.Original.Decision,values=d.Values??[];
  assert.deepEqual(p.branches.map(b=>b.origin),[-1,...(d.Pool??[])]);assert.deepEqual(t.branches.map(b=>b.origin),p.branches.map(b=>b.origin));
  for(const v of values)near(queryMixtureScore(v.Conditional,v.Mass[1]),v.Gain);
  assert.equal(chooseShrinkage(values,1).origin,d.Selected[3]);
  const rows=values.map(v=>({origin:v.Origin,p:v.Mass[1],y:Number(source.Steps[v.Origin].Y),q:source.Steps[v.Origin].Q,weight:1/values.length}));
  runs.push({phase:p.phase,case:p.case,index:p.index,schedule:p.schedule,values,rows,p,t,controls:d.Selected});
}
assert.equal(runs.length,2688);const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],pop.artifactHashes[3]);assert.equal(hashes[0],sample.artifactHashes[4]);assert.equal(hashes[1],pop.artifactHashes[1]);assert.equal(hashes[1],sample.artifactHashes[1]);assert.equal(headers[0].InputSHA256,hashes[1]);
const train=(rs,teacher=false)=>fitQueryShrinkage(rs.filter(r=>r.phase===0&&r.schedule===1).flatMap(r=>r.rows.map(x=>({...x,y:teacher?x.q:x.y}))));
const labelModel=train(runs),teacherModel=train(runs,true);assert.equal(labelModel.rows,4640);assert(Math.abs(labelModel.weight-672)<1e-8);
const poisoned=runs.map(r=>r.phase===0?r:{...r,rows:r.rows.map(x=>({...x,y:1-x.y,q:1-x.q}))});assert.deepEqual(train(poisoned),labelModel);assert.deepEqual(train(poisoned,true),teacherModel);
const records=[];
for(const r of runs){
  const label=chooseShrinkage(r.values,labelModel.alpha),teacherFit=chooseShrinkage(r.values,teacherModel.alpha),neutral=chooseShrinkage(r.values,0);
  const teacherValues=r.values.map((v,j)=>({...v,Mass:[1-r.rows[j].q,r.rows[j].q]})),oracle=chooseShrinkage(teacherValues,1);
  const selected=[...r.controls,label.origin,teacherFit.origin,neutral.origin,oracle.origin];
  const pb=new Map(r.p.branches.map(b=>[b.origin,b])),sb=new Map(r.t.branches.map(b=>[b.origin,b]));
  const actual=selected.map(j=>sb.get(j).actual),expected=selected.map(j=>sb.get(j).value),integrated=selected.map(j=>pb.get(j).population),costs=selected.map(j=>Number(j>=0));
  for(let j=0;j<4;j++){near(integrated[j],r.p.population[j]);near(expected[j],r.t.values[j]);}
  if(r.schedule===0){assert(selected.every(j=>j===-1));for(const xs of [actual,expected,integrated])xs.forEach(x=>near(x,xs[0]));}
  else assert(costs.slice(1).every(c=>c===1));
  const massMSE=r.rows.length?[1,labelModel.alpha,teacherModel.alpha,0].map(a=>mean(r.rows.map(x=>(shrinkQueryMass(x.p,a)-x.q)**2))):null;
  records.push({phase:r.phase,case:r.case,index:r.index,schedule:r.schedule,selected,actual,expected,integrated,costs,massMSE,labelScores:label.scores});
}
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const summarize=rs=>{
  const out={records:rs.length,costs:Array.from({length:8},(_,a)=>rs.reduce((s,r)=>s+r.costs[a],0))};
  for(const field of ['actual','expected','integrated'])out[field]=Array.from({length:8},(_,a)=>mean(rs.map(r=>r[field][a])));
  out.massMSE=rs.some(r=>r.massMSE)?Array.from({length:4},(_,a)=>mean(rs.filter(r=>r.massMSE).map(r=>r.massMSE[a]))):null;return out;
};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  const gains={};for(const field of ['actual','expected','integrated'])gains[field]=[4,5,6,7].map(a=>[1,2].map(control=>ci(rs.map(r=>r[field][control]-r[field][a]))));
  groups.push({phase,case:c,schedule,...summarize(rs),gains});
}
const evaluation=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const field of ['actual','expected','integrated'])screens[field]=[0,1,2,3].map(a=>{
  const nonharm=evaluation.every(g=>g.gains[field][a].every(x=>x.lower>=-.001)),transition=evaluation.filter(g=>[19,20].includes(g.case)).every(g=>g.gains[field][a].every(x=>x.lower>0));
  return {arm:a+4,nonharm,transition,pass:nonharm&&transition,nonharmCells:evaluation.filter(g=>g.gains[field][a].every(x=>x.lower>=-.001)).length,positiveCells:evaluation.filter(g=>g.gains[field][a].every(x=>x.lower>0)).length};
});
console.log(JSON.stringify({scope:'Frozen single-parameter mixture-score heuristic on consumed data; not a proper-risk certificate or untouched confirmation',arms:['noquery','random','entropy','joint','label-fit','teacher-fit-diagnostic','neutral','teacher-direct-oracle'],hashes:Object.fromEntries([[rawPath,hashes[0]],[sourcePath,hashes[1]],...[popPath,samplePath,'research/query-shrinkage.mjs','research/query-shrinkage-test.mjs','research/query-shrinkage-experiment.mjs','research/regime-query-outcome-audit.mjs','docs/experiments/mmm-query-shrinkage-protocol.md'].map(p=>[p,sha(p)])]),labelModel,teacherModel,screens,phase1Delayed:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),groups,records}));
const evalRuns=runs.filter(r=>r.phase===1&&r.schedule===1),times=[];for(let i=0;i<100;i++)chooseShrinkage(evalRuns[i].values,labelModel.alpha);
for(let repeat=0;repeat<3;repeat++)for(const r of evalRuns){const t=performance.now();chooseShrinkage(r.values,labelModel.alpha);times.push(performance.now()-t);}
times.sort((a,b)=>a-b);const q=p=>times[Math.ceil(p*times.length)-1];
console.error(JSON.stringify({scope:'Selection with precomputed Bayesian conditional outputs; excludes posterior fits, retrieval, I/O and serving',node:process.version,calls:times.length,selectionMS:{median:q(.5),p95:q(.95),p99:q(.99),max:times.at(-1)},scriptSHA256:sha(import.meta.filename)}));
