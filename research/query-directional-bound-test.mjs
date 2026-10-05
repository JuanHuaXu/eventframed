import assert from 'node:assert/strict';
import {directionalGainBound} from './query-directional-bound.mjs';
const grid=[0,.25,.5,.75,1],near=(a,b)=>assert(Math.abs(a-b)<1e-12);let checks=0;
const boxes=[[[0,1],[0,1],[0,1]],[[.25,.75],[0,.5],[.5,1]],[[.5,.5],[.25,.25],[.75,.75]]];
const loss=(p,q)=>q*(1-p)**2+(1-q)*p*p;
for(const b of grid)for(const f0 of grid)for(const f1 of grid)for(const box of boxes){
  const snapshot=structuredClone(box),r=directionalGainBound(b,[f0,f1],box[0],[box[1],box[2]]),seen=[];
  for(const a of grid)for(const c of grid)for(const d of grid){
    const p=box[0][0]+a*(box[0][1]-box[0][0]),q0=box[1][0]+c*(box[1][1]-box[1][0]),q1=box[2][0]+d*(box[2][1]-box[2][0]);
    const gain=(1-p)*(loss(b,q0)-loss(f0,q0))+p*(loss(b,q1)-loss(f1,q1));
    assert(gain>=r.lower-1e-12&&gain<=r.upper+1e-12);seen.push(gain);checks++;
  }
  near(Math.min(...seen),r.lower);near(Math.max(...seen),r.upper);assert.deepEqual(box,snapshot);
  const flip=directionalGainBound(b,[f1,f0],[1-box[0][1],1-box[0][0]],[box[2],box[1]]);near(flip.lower,r.lower);near(flip.upper,r.upper);
}
const good=directionalGainBound(.8,[.4,.4],[0,1],[[.5,.5],[.5,.5]]);near(good.lower,.08);
const bad=directionalGainBound(.5,[.1,.9],[0,1],[[.5,.5],[.5,.5]]);assert(bad.upper<0);
// Genuine predictive dependence can help even though the unchanged marginal
// remains .5; no independent-query/target assumption is built into this bound.
const dependent=directionalGainBound(.5,[.1,.9],[.5,.5],[[.1,.1],[.9,.9]]);near(dependent.lower,.16);
for(const p of [NaN,Infinity,-.1,1.1])assert.throws(()=>directionalGainBound(p,[.2,.8],[0,1],[[0,1],[0,1]]));
assert.throws(()=>directionalGainBound(.5,[.2,.8],[.9,.1],[[0,1],[0,1]]));
console.log(JSON.stringify({status:'PASS',interiorChecks:checks,boxes:375,scope:'Independent direct Brier enumeration, exact corner extrema, relabeling, ownership, invalid input and dependent-target control; interval coverage is not established'}));
