import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {criticFeatures} from './query-critic.mjs';
import {fitQueryFactorial,chooseQueryFactorial as chooseCritic} from './query-error-factorial.mjs';

const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const targetPath=process.argv[2],rawPath=process.argv[3];
const targets=JSON.parse(fs.readFileSync(targetPath,'utf8'));assert.equal(targets.records.length,2688);
const stream=fs.createReadStream(rawPath),rawHash=crypto.createHash('sha256');stream.on('data',b=>rawHash.update(b));
const it=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
const header=JSON.parse((await it.next()).value);assert.equal(header.Version,'regime-query-disjoint-v1');assert.equal(header.InputSHA256,targets.artifactHashes[1]);
const projection=JSON.parse(fs.readFileSync(process.argv[5],'utf8')),variant=Number(process.argv[6]);assert([0,1,2,3].includes(variant));assert.equal(projection.sourceSHA256,targets.artifactHashes[1]);assert.equal(projection.records.length,2688);
const augmented=variant>=2,quadratic=variant%2===1;
const runs=[],ids=new Set();let index=0;
for(;;){
  const line=await it.next();if(line.done)break;const r=JSON.parse(line.value),t=targets.records[index++];assert(t&&!r.Error);
  const old=r.Original,newer=r.Candidate,d=old.Decision;
  for(const [key,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(old[key],t[lower]);assert.equal(newer[key],t[lower]);}
  const id=`${t.phase}:${t.case}:${t.index}:${t.schedule}`;assert(!ids.has(id));ids.add(id);
  assert.deepEqual(d.Pool,newer.Decision.Pool);
  assert.deepEqual(t.branches.map(b=>b.origin),[-1,...(d.Pool??[])]);
  const projected=projection.records[index-1];for(const k of ['phase','case','index','schedule'])assert.equal(projected[k],t[k]);assert.deepEqual(projected.features.map(c=>c.origin),d.Pool??[]);
  const candidates=(d.Pool??[]).map(origin=>{
    const x=criticFeatures(d,newer.Decision,origin),b=t.branches.find(b=>b.origin===origin);
    return {origin,x:augmented?x.concat(projected.features.find(c=>c.origin===origin).values):x,y:t.branches[0].value-b.value,weight:1/d.Pool.length};
  });
  runs.push({phase:t.phase,case:t.case,index:t.index,schedule:t.schedule,candidates,targets:t.branches,controls:[...d.Selected,newer.Decision.Selected[3]]});
}
assert.equal(index,2688);const rawSHA256=rawHash.digest('hex');assert.equal(rawSHA256,targets.artifactHashes[4]);
const train=rs=>fitQueryFactorial(rs.filter(r=>r.phase===0&&r.schedule===1).map(r=>r.candidates),quadratic);
const model=train(runs);assert(Math.abs(model.context.weight-672)<1e-8);
// Test targets, outcomes and oracle quantities cannot affect fit or selection.
const poisoned=runs.map(r=>r.phase===0?r:{...r,candidates:r.candidates.map(c=>({...c,y:1-c.y})),targets:r.targets.map(b=>({...b,value:1-b.value,actual:1-b.actual,counter:1-b.counter,q:1-(b.q??0),y:!b.y}))});
assert.deepEqual(train(poisoned),model);
const baseline=JSON.parse(fs.readFileSync(process.argv[4],'utf8'));assert.equal(baseline.rawSHA256,rawSHA256);assert.equal(baseline.targetSHA256,sha(targetPath));if(variant===0){assert.deepEqual(model.context,baseline.model.context);assert.deepEqual(model.rank,baseline.model.rank);}
const records=[];
for(let i=0;i<runs.length;i++){
  const r=runs[i],decision=chooseCritic(model,r.candidates);assert.deepEqual(chooseCritic(model,poisoned[i].candidates),decision);if(variant===0)assert.deepEqual(decision,baseline.records[i].decision);
  const selected=[...r.controls,decision.forced,decision.gated,decision.gated<0?-1:r.controls[1],decision.gated<0?-1:r.controls[2]];
  const branch=j=>{const b=r.targets.find(b=>b.origin===j);assert(b);return b;};
  const actual=selected.map(j=>branch(j).actual),expected=selected.map(j=>branch(j).value),costs=selected.map(j=>Number(j>=0));
  assert.equal(costs[6],costs[7]);assert.equal(costs[6],costs[8]);assert.equal(costs[5],costs[1]);assert.equal(costs[5],costs[2]);
  if(r.schedule===0){assert(selected.every(j=>j===-1));assert(actual.every(x=>x===actual[0]));}
  records.push({phase:r.phase,case:r.case,index:r.index,schedule:r.schedule,selected,actual,expected,costs,decision});
}
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  groups.push({phase,case:c,schedule,actual:Array.from({length:9},(_,a)=>mean(rs.map(r=>r.actual[a]))),expected:Array.from({length:9},(_,a)=>mean(rs.map(r=>r.expected[a]))),costs:Array.from({length:9},(_,a)=>rs.reduce((s,r)=>s+r.costs[a],0)),
    forcedGains:[1,2].map(a=>ci(rs.map(r=>r.actual[a]-r.actual[5]))),gatedGains:[7,8].map(a=>ci(rs.map(r=>r.actual[a]-r.actual[6])))});
}
const evaluation=groups.filter(g=>g.phase===1&&g.schedule===1);
const screen=field=>{const nonharm=evaluation.every(g=>g[field].every(v=>v.lower>=-.001));const transition=evaluation.filter(g=>[19,20].includes(g.case)).every(g=>g[field].every(v=>v.lower>0));return {nonharm,transition,pass:nonharm&&transition,nonharmCells:evaluation.filter(g=>g[field].every(v=>v.lower>=-.001)).length,positiveCells:evaluation.filter(g=>g[field].every(v=>v.lower>0)).length};};
const delayed=records.filter(r=>r.phase===1&&r.schedule===1);
for(let i=0;i<records.length;i++)for(const key of ['selected','actual','expected','costs'])assert.deepEqual(records[i][key].slice(0,5),baseline.records[i][key].slice(0,5));
const hashes=Object.fromEntries(['research/query-critic.mjs','research/query-critic-test.mjs','research/residual-logit.mjs','research/query-error-factorial.mjs','research/query-error-factorial-test.mjs','research/query-error-factorial-experiment.mjs','research/prequential-projection.mjs','docs/experiments/mmm-query-error-factorial-protocol.md','docs/experiments/mmm-query-centered-critic-protocol.md','docs/experiments/mmm-query-critic-protocol.md'].map(p=>[p,sha(p)]));
console.log(JSON.stringify({scope:'Frozen state-information by polynomial-degree factorial, consumed phase1 on consumed data; not untouched confirmation or whole-goal success',
  arms:['noquery','random','entropy','joint','disjoint','critic-forced','critic-gated','random-same-gate','entropy-same-gate'],
  variant,augmented,quadratic,projectionSHA256:sha(process.argv[5]),baselineSHA256:sha(process.argv[4]),targetSHA256:sha(targetPath),rawSHA256,hashes,model,modelSHA256:crypto.createHash('sha256').update(JSON.stringify(model)).digest('hex'),
  phase1Delayed:{records:delayed.length,actual:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.actual[a]))),expected:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.expected[a]))),costs:Array.from({length:9},(_,a)=>delayed.reduce((s,r)=>s+r.costs[a],0))},
  forcedScreen:screen('forcedGains'),gatedScreen:screen('gatedGains'),groups,records}));
