import assert from 'node:assert/strict';
import {acquisitionBelief,acquisitionChoice,TYPES} from './interval-acquisition.mjs';
import {lookaheadChoice} from './interval-lookahead.mjs';
import {countState,canonicalHistory,createCountPlanner} from './count-planning.mjs';
let checks=0,planningChecks=0,maxError=0;
for(const roots of [0,1,3,5,10,15])for(const counts of [[0,0,0,0,0,0,0,0],[1,1,0,1,0,0,0,0],[1,0,0,1,1,0,0,2]]){
  const h=canonicalHistory(roots,counts),reversed=h.slice(0,4),slots=[0,0,0,0];
  for(const s of h.slice(4).reverse()){const j=TYPES.indexOf(s.action[1]);reversed.push({action:[1,TYPES[j],slots[j]++],outcome:s.outcome});}
  assert.deepEqual(countState(reversed),{roots,counts});
  const a=acquisitionBelief(h),b=acquisitionBelief(reversed);
  for(const e of [Math.abs(a.mass-b.mass),...a.forecast.map((p,i)=>Math.abs(p-b.forecast[i]))]){assert(e<1e-12);maxError=Math.max(maxError,e);checks++;}
  const planner=createCountPlanner(roots);
  const choices=[acquisitionChoice(h,'target',()=>{throw Error('rng');}),lookaheadChoice(h)];
  choices.forEach((action,i)=>{const p=planner.plan(counts,i+1);assert(p.costs[TYPES.indexOf(action[1])]<=p.cost+1e-12);planningChecks++;});
}
assert.throws(()=>canonicalHistory(0,[7,0,0,0,0,0,0,0]));
console.log(JSON.stringify({exchangeabilityChecks:checks,maxError,planningChecks,negativeTests:1}));

