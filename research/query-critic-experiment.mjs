import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {criticFeatures,fitCritic,chooseCritic} from './query-critic.mjs';

const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const targetPath=process.argv[2],rawPath=process.argv[3];
const targets=JSON.parse(fs.readFileSync(targetPath,'utf8'));assert.equal(targets.records.length,2688);
const stream=fs.createReadStream(rawPath),rawHash=crypto.createHash('sha256');stream.on('data',b=>rawHash.update(b));
const it=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
const header=JSON.parse((await it.next()).value);assert.equal(header.Version,'regime-query-disjoint-v1');assert.equal(header.InputSHA256,targets.artifactHashes[1]);
const runs=[],ids=new Set();let index=0;
for(;;){
  const line=await it.next();if(line.done)break;const r=JSON.parse(line.value),t=targets.records[index++];assert(t&&!r.Error);
  const old=r.Original,newer=r.Candidate,d=old.Decision;
  for(const [key,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(old[key],t[lower]);assert.equal(newer[key],t[lower]);}
  const id=`${t.phase}:${t.case}:${t.index}:${t.schedule}`;assert(!ids.has(id));ids.add(id);
  assert.deepEqual(d.Pool,newer.Decision.Pool);
  assert.deepEqual(t.branches.map(b=>b.origin),[-1,...(d.Pool??[])]);
  const candidates=(d.Pool??[]).map(origin=>{
    const x=criticFeatures(d,newer.Decision,origin),b=t.branches.find(b=>b.origin===origin);
    return {origin,x,y:t.branches[0].value-b.value,weight:1/d.Pool.length};
  });
  runs.push({phase:t.phase,case:t.case,index:t.index,schedule:t.schedule,candidates,targets:t.branches,controls:[...d.Selected,newer.Decision.Selected[3]]});
}
assert.equal(index,2688);const rawSHA256=rawHash.digest('hex');assert.equal(rawSHA256,targets.artifactHashes[4]);
const train=rs=>fitCritic(rs.filter(r=>r.phase===0&&r.schedule===1).flatMap(r=>r.candidates));
const model=train(runs);assert(Math.abs(model.weight-672)<1e-8);
// Test targets, outcomes and oracle quantities cannot affect fit or selection.
const poisoned=runs.map(r=>r.phase===0?r:{...r,candidates:r.candidates.map(c=>({...c,y:1-c.y})),targets:r.targets.map(b=>({...b,value:1-b.value,actual:1-b.actual,counter:1-b.counter,q:1-(b.q??0),y:!b.y}))});
assert.deepEqual(train(poisoned),model);
const records=[];
for(let i=0;i<runs.length;i++){
  const r=runs[i],decision=chooseCritic(model,r.candidates);assert.deepEqual(chooseCritic(model,poisoned[i].candidates),decision);
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
const hashes=Object.fromEntries(['research/query-critic.mjs','research/query-critic-test.mjs','research/residual-logit.mjs','research/query-critic-experiment.mjs','docs/experiments/mmm-query-critic-protocol.md'].map(p=>[p,sha(p)]));
console.log(JSON.stringify({scope:'Phase0-trained, frozen phase1 one-decision critic on consumed data; not untouched confirmation or whole-goal success',
  arms:['noquery','random','entropy','joint','disjoint','critic-forced','critic-gated','random-same-gate','entropy-same-gate'],
  targetSHA256:sha(targetPath),rawSHA256,hashes,model,modelSHA256:crypto.createHash('sha256').update(JSON.stringify(model)).digest('hex'),
  phase1Delayed:{records:delayed.length,actual:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.actual[a]))),expected:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.expected[a]))),costs:Array.from({length:9},(_,a)=>delayed.reduce((s,r)=>s+r.costs[a],0))},
  forcedScreen:screen('forcedGains'),gatedScreen:screen('gatedGains'),groups,records}));
