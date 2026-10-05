import assert from 'node:assert/strict';
import {projectedTransfer,simplexProjection} from './projected-transfer.mjs';
import {guardedTransfer} from './robust-transfer.mjs';
let seed=17191;const random=()=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return (seed+.5)/4294967296;};
const simplex=()=>{const a=Array.from({length:4},random),s=a.reduce((x,y)=>x+y,0);return a.map(x=>x/s);};
let maxError=0,maxCycles=0,maxGap=0;
for(let i=0;i<300;i++){
  const b=random(),p=random(),qs=Array.from({length:8},random),epsilon=.01;
  let lo=0,hi=1;
  for(const q of qs){const radius=Math.sqrt((b-q)**2+epsilon/2);lo=Math.max(lo,q-radius);hi=Math.min(hi,q+radius);}
  const exact=Math.max(lo,Math.min(hi,p)),r=projectedTransfer([b,1-b],[p,1-p],qs.map(q=>[q,1-q]),epsilon);
  maxError=Math.max(maxError,Math.abs(exact-r.forecast[0]));assert(Math.abs(exact-r.forecast[0])<2e-6);maxCycles=Math.max(maxCycles,r.cycles);maxGap=Math.max(maxGap,r.gap);
}
let variationalChecks=0;
for(let i=0;i<100;i++){
  const base=simplex(),proposal=simplex(),laws=Array.from({length:8},simplex),r=projectedTransfer(base,proposal,laws);
  maxCycles=Math.max(maxCycles,r.cycles);maxGap=Math.max(maxGap,r.gap);
  for(let j=0;j<20;j++){
    const feasible=guardedTransfer(base,simplex(),laws).forecast;
    const violation=proposal.reduce((s,x,k)=>s+(x-r.forecast[k])*(feasible[k]-r.forecast[k]),0);
    assert(violation<3e-6);variationalChecks++;
  }
}
assert.deepEqual(simplexProjection([2,-1]),[1,0]);
assert.equal(projectedTransfer([.5,.5],[.5,.5],[[1,0]]).cycles,0);
console.log(JSON.stringify({analyticChecks:300,maxError,variationalChecks,maxCycles,maxGap}));
