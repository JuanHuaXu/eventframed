import assert from 'node:assert/strict';
import {weightedRisk,minimumMassRisk} from './query-mass.mjs';
const values=[0,.25,.5,.75,1];let cases=0;
for(const p of values)for(const q of values)for(const a of values)for(const b of values){
  assert(Math.abs(weightedRisk([a,b],p)-weightedRisk([a,b],q)-(p-q)*(b-a))<1e-15);
  assert.equal(weightedRisk([a,b],p),weightedRisk([b,a],1-p));cases++;
}
let boundCases=0;
for(const p of values)for(const q of values)for(const a of values)for(const b of values)for(const c of values){
  const actions=[{origin:152,losses:[a,b],probability:p},{origin:153,losses:[c,1-c],probability:1-p}];
  const truth=[{...actions[0],probability:q},{...actions[1],probability:1-q}],mp=minimumMassRisk(actions),mq=minimumMassRisk(truth);
  const risk=(rows,j)=>{const r=rows.find(x=>x.origin===j);return weightedRisk(r.losses,r.probability);};
  const regret=risk(truth,mp.origin)-mq.risk,bound=Math.abs(risk(actions,mp.origin)-risk(truth,mp.origin))+Math.abs(risk(actions,mq.origin)-risk(truth,mq.origin));
  assert(regret>=-1e-15&&regret<=bound+1e-15);assert.deepEqual(minimumMassRisk(actions.toReversed()),mp);boundCases++;
}
const tie=[{origin:152,losses:[.2,.2],probability:.1},{origin:-1,losses:[.2,.2],probability:.9}],copy=structuredClone(tie);
assert.equal(minimumMassRisk(tie).origin,-1);assert.deepEqual(tie,copy);
for(const a of tie)for(const k of ['teacher','q','actualY','future'])Object.defineProperty(a,k,{get(){throw Error(k);}});
assert.equal(minimumMassRisk(tie).origin,-1);
for(const p of [NaN,Infinity,-.1,1.1])assert.throws(()=>weightedRisk([0,1],p));
for(const losses of [[],[0],[0,1,2],[-1,1],[0,Infinity]])assert.throws(()=>weightedRisk(losses,.5));
assert.throws(()=>minimumMassRisk([]));assert.throws(()=>minimumMassRisk([copy[0],copy[0]]));
console.log(`PASS: ${cases} identities/relabelings,${boundCases} regret/order controls, ownership, ties, oracle traps and invalid inputs`);
