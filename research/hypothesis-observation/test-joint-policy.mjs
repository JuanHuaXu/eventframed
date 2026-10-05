import assert from 'node:assert/strict';
import {brierOptimum} from './joint-policy.mjs';
let checks=0,maxError=0;
for(let k=0;k<64;k++){
  const joint=Array.from({length:4},(_,i)=>((i+3)*17+k*13)%41/47),m=joint.reduce((a,v)=>a+v,0),out=brierOptimum(joint);
  assert(out.forecast.every(p=>p>=0&&p<=1));assert(Math.abs(out.forecast.reduce((a,p)=>a+p,0)-1)<1e-12);
  for(let j=0;j<16;j++){
    const raw=Array.from({length:4},(_,i)=>1+((i+1)*19+j*7)%23),sum=raw.reduce((a,v)=>a+v,0),p=raw.map(v=>v/sum);
    const loss=joint.reduce((a,v,y)=>a+v*p.reduce((b,q,z)=>b+(q-Number(y===z))**2,0),0);
    const excess=m*p.reduce((a,q,y)=>a+(q-out.forecast[y])**2,0),error=Math.abs(loss-out.cost-excess);
    assert(error<1e-12);assert(loss>=out.cost-1e-12);maxError=Math.max(maxError,error);checks++;
  }
}
assert.deepEqual(brierOptimum([0,0,0,0]),{forecast:[.25,.25,.25,.25],cost:0});
assert.deepEqual(brierOptimum([0,2,0,0]),{forecast:[0,1,0,0],cost:0});
for(const bad of [[NaN,0,0,0],[-1,1,0,0],[1,2],[Infinity,0,0,0]])assert.throws(()=>brierOptimum(bad));
console.log(JSON.stringify({excessRiskChecks:checks,maxError,boundaryTests:2,invalidTests:4}));
