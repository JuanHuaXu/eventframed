import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {prepare,bestResponse} from './global-policy.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'global-policy-game.json'),d=JSON.parse(raw);
const source=JSON.parse(readFileSync(dir+'risk-budget-evaluation.json')),model=JSON.parse(readFileSync(dir+'count-planning-exact.json'));
const worlds=source.results.map(r=>({noise:r.noise,mask:r.mask})),graphs=prepare(model,worlds);
const witness=d.rounds.reduce((a,b)=>a.lower>b.lower?a:b),results=[];
// Remove a control from an existing dual distribution, then recompute the oracle.
// No claim that this projected witness is the strongest possible lower bound.
for(const controls of [['random','entropy'],['random','tie_entropy']]){
  const q=witness.weights.map((v,i)=>controls.includes(d.rows[i].control)?v:0),sum=q.reduce((a,v)=>a+v,0);
  assert(sum>0);q.forEach((v,i)=>q[i]=v/sum);
  const fw=Array(80).fill(0),aw=Array(80).fill(0);let constant=0;
  d.rows.forEach((r,i)=>{(r.metric==='final'?fw:aw)[r.g]+=q[i];constant+=q[i]*r.threshold;});
  const response=bestResponse(graphs,fw,aw);
  results.push({controls,retainedWeight:sum,lower:response.objective-constant,forwardError:response.forwardError,weights:q});
}
console.log(JSON.stringify({scope:'Projected dual witnesses for each two-control screen; numerical, not optimality or interval proof',inputSHA256:createHash('sha256').update(raw).digest('hex'),sourceSHA256:createHash('sha256').update(readFileSync(dir+'global-policy-reduced-witness.mjs')).digest('hex'),results},null,2));
