import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';
import {criticFeatures} from './query-critic.mjs';
import {fitQueryFactorial,chooseQueryFactorial} from './query-error-factorial.mjs';

const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const read=p=>JSON.parse(fs.readFileSync(p,'utf8'));
const [populationPath,samplePath,rawPath,projectionPath,baselinePath,variantText]=process.argv.slice(2);
const population=read(populationPath),sample=read(samplePath),projection=read(projectionPath),baseline=read(baselinePath),variant=Number(variantText);
assert([0,1,2,3].includes(variant));assert.equal(baseline.variant,variant);
const augmented=variant>=2,quadratic=variant%2===1,source=population.artifactHashes[1];
assert.equal(source,sample.artifactHashes[1]);assert.equal(source,projection.sourceSHA256);
assert.equal(sha(samplePath),baseline.targetSHA256);assert.equal(sha(projectionPath),baseline.projectionSHA256);
for(const obj of [population,sample,projection,baseline])assert.equal(obj.records.length,2688);
const stream=fs.createReadStream(rawPath),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));
const it=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
const header=JSON.parse((await it.next()).value);assert.equal(header.Version,'regime-query-disjoint-v1');assert.equal(header.InputSHA256,source);
const runs=[];
for(;;){
  const line=await it.next();if(line.done)break;
  const raw=JSON.parse(line.value),i=runs.length,p=population.records[i],s=sample.records[i],f=projection.records[i],b=baseline.records[i];assert(p&&!raw.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){
    for(const r of [s,f,b])assert.equal(p[lower],r[lower]);
    for(const r of [raw.Original,raw.Candidate])assert.equal(p[lower],r[upper]);
  }
  const d=raw.Original.Decision,n=raw.Candidate.Decision,pool=d.Pool??[];
  assert.deepEqual(pool,n.Pool??[]);assert.deepEqual(p.branches.map(x=>x.origin),[-1,...pool]);assert.deepEqual(s.branches.map(x=>x.origin),[-1,...pool]);assert.deepEqual(f.features.map(x=>x.origin),pool);
  for(let j=0;j<p.branches.length;j++)assert(Math.abs(p.branches[j].sample-s.branches[j].value)<1e-12);
  const candidates=pool.map(origin=>{
    const x=criticFeatures(d,n,origin),branch=p.branches.find(x=>x.origin===origin),old=s.branches.find(x=>x.origin===origin);
    return {origin,x:augmented?x.concat(f.features.find(x=>x.origin===origin).values):x,
      y:p.branches[0].population-branch.population,oldY:s.branches[0].value-old.value,weight:1/pool.length};
  });
  runs.push({phase:p.phase,case:p.case,index:p.index,schedule:p.schedule,candidates,p,s,controls:[...d.Selected,n.Selected[3]]});
}
assert.equal(runs.length,2688);assert.equal(new Set(runs.map(r=>`${r.phase}:${r.case}:${r.index}:${r.schedule}`)).size,2688);
const rawSHA256=hash.digest('hex');assert.equal(rawSHA256,baseline.rawSHA256);assert.equal(rawSHA256,population.artifactHashes[3]);assert.equal(rawSHA256,sample.artifactHashes[4]);
const train=(rs,old=false)=>fitQueryFactorial(rs.filter(r=>r.phase===0&&r.schedule===1).map(r=>r.candidates.map(c=>({...c,y:old?c.oldY:c.y}))),quadratic);
const model=train(runs),oldModel=train(runs,true);assert.deepEqual(oldModel,baseline.model);
assert.equal(model.context.rows,4640);assert(Math.abs(model.context.weight-672)<1e-8);
const poisoned=runs.map(r=>r.phase===0?r:{...r,candidates:r.candidates.map(c=>({...c,y:1-c.y,oldY:1-c.oldY}))});assert.deepEqual(train(poisoned),model);
const records=[];
for(let i=0;i<runs.length;i++){
  const r=runs[i],decision=chooseQueryFactorial(model,r.candidates),oldDecision=chooseQueryFactorial(oldModel,r.candidates);
  assert.deepEqual(oldDecision,baseline.records[i].decision);assert.deepEqual(chooseQueryFactorial(model,poisoned[i].candidates),decision);
  // The selector receives only predecision features and origins, never targets.
  const inference=r.candidates.map(c=>({origin:c.origin,x:c.x}));
  for(const c of inference)for(const key of ['y','oldY','q','teacher','phase','case','future'])Object.defineProperty(c,key,{get(){throw Error(`oracle access:${key}`);}});
  assert.deepEqual(chooseQueryFactorial(model,inference),decision);
  const selected=[...r.controls,decision.forced,decision.gated,decision.gated<0?-1:r.controls[1],decision.gated<0?-1:r.controls[2],oldDecision.forced,oldDecision.gated];
  const branch=(rs,j)=>{const b=rs.find(b=>b.origin===j);assert(b);return b;};
  const actual=selected.map(j=>branch(r.s.branches,j).actual),expected=selected.map(j=>branch(r.s.branches,j).value),integrated=selected.map(j=>branch(r.p.branches,j).population),costs=selected.map(j=>Number(j>=0));
  for(const [key,values]of [['actual',actual],['expected',expected],['costs',costs]]){
    assert.deepEqual(values.slice(0,5),baseline.records[i][key].slice(0,5));assert.deepEqual(values.slice(9),baseline.records[i][key].slice(5,7));
  }
  assert.equal(costs[6],costs[7]);assert.equal(costs[6],costs[8]);assert.equal(costs[5],costs[1]);assert.equal(costs[5],costs[2]);
  if(r.schedule===0){assert(selected.every(j=>j===-1));for(const values of [actual,expected,integrated])assert(values.every(x=>x===values[0]));}
  records.push({phase:r.phase,case:r.case,index:r.index,schedule:r.schedule,selected,actual,expected,integrated,costs,decision});
}
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  const g={phase,case:c,schedule,costs:Array.from({length:11},(_,a)=>rs.reduce((s,r)=>s+r.costs[a],0))};
  for(const field of ['actual','expected','integrated'])g[field]={means:Array.from({length:11},(_,a)=>mean(rs.map(r=>r[field][a]))),forcedGains:[1,2].map(a=>ci(rs.map(r=>r[field][a]-r[field][5]))),gatedGains:[7,8].map(a=>ci(rs.map(r=>r[field][a]-r[field][6])))};
  groups.push(g);
}
const evaluation=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const field of ['actual','expected','integrated']){
  screens[field]={};for(const mode of ['forced','gated']){const k=mode+'Gains',nonharm=evaluation.every(g=>g[field][k].every(v=>v.lower>=-.001)),transition=evaluation.filter(g=>[19,20].includes(g.case)).every(g=>g[field][k].every(v=>v.lower>0));screens[field][mode]={nonharm,transition,pass:nonharm&&transition,nonharmCells:evaluation.filter(g=>g[field][k].every(v=>v.lower>=-.001)).length,positiveCells:evaluation.filter(g=>g[field][k].every(v=>v.lower>0)).length};}
}
const delayed=records.filter(r=>r.phase===1&&r.schedule===1),phase1Delayed={records:delayed.length,costs:Array.from({length:11},(_,a)=>delayed.reduce((s,r)=>s+r.costs[a],0))};
for(const field of ['actual','expected','integrated'])phase1Delayed[field]=Array.from({length:11},(_,a)=>mean(delayed.map(r=>r[field][a])));
const output={scope:'Population-target-only rescue on consumed data; not untouched confirmation',variant,augmented,quadratic,
  arms:['noquery','random','entropy','joint','disjoint','critic-forced','critic-gated','random-same-gate','entropy-same-gate','previous-forced','previous-gated'],
  hashes:Object.fromEntries([populationPath,samplePath,rawPath,projectionPath,baselinePath,'research/population-critic-experiment.mjs','research/query-error-factorial.mjs','research/query-critic.mjs','research/residual-logit.mjs','docs/experiments/mmm-population-critic-protocol.md'].map(p=>[p,sha(p)])),model,screens,phase1Delayed,groups,records};
console.log(JSON.stringify(output));
// Timing is emitted separately so deterministic replay remains byte-comparable.
const fitMS=[];for(let i=0;i<3;i++){const t=performance.now(),m=train(runs);fitMS.push(performance.now()-t);assert.deepEqual(m,model);}
const evalRuns=runs.filter(r=>r.phase===1&&r.schedule===1),times=[];
for(let i=0;i<100;i++)chooseQueryFactorial(model,evalRuns[i].candidates);
for(let repeat=0;repeat<3;repeat++)for(const r of evalRuns){const t=performance.now();chooseQueryFactorial(model,r.candidates);times.push(performance.now()-t);}
times.sort((a,b)=>a-b);const q=p=>times[Math.ceil(p*times.length)-1];
console.error(JSON.stringify({scope:'Component timing with precomputed feature vectors; excludes Bayesian/projection feature creation, retrieval, I/O and serving',variant,node:process.version,fitMS,calls:times.length,selectionMS:{median:q(.5),p95:q(.95),p99:q(.99),max:times.at(-1)},scriptSHA256:sha(import.meta.filename)}));
