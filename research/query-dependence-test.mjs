import assert from 'node:assert/strict';
import {binaryJoint,fitDependence} from './query-dependence.mjs';
const near=(a,b)=>assert(Math.abs(a-b)<1e-11);
let cases=0;
for(const p of [.01,.2,.5,.9,.99])for(const f0 of [.02,.3,.8])for(const f1 of [.01,.4,.97])for(const lambda of [0,.2,.7,1]){
  const r=binaryJoint(p,f0,f1,lambda),[a,b,c,d]=r.joint;
  near(a+b+c+d,1);near(c+d,p);near(b+d,r.base);assert(r.joint.every(x=>x>0));
  const flip=binaryJoint(1-p,f1,f0,lambda);r.joint.forEach((v,i)=>near(v,flip.joint[(i+2)%4]));
  const targetFlip=binaryJoint(p,1-f0,1-f1,lambda);r.joint.forEach((v,i)=>near(v,targetFlip.joint[i^1]));
  near(p*(1-p)*(r.conditional[1]-r.conditional[0])**2,lambda**2*p*(1-p)*(f1-f0)**2);cases++;
}
for(const truth of [0,.17,.5,.83,1]){
  const rows=[];
  for(const query of [0,1]){
    const base=.5,delta=query?.3:-.3,q=base+truth*delta;
    for(const y of [0,1])rows.push({base,delta,y,weight:.5*(y?q:1-q)});
  }
  const original=structuredClone(rows);near(fitDependence(rows).lambda,truth);assert.deepEqual(rows,original);
  for(const row of rows)for(const k of ['teacher','phase1','case','q'])Object.defineProperty(row,k,{get(){throw Error(k);}});
  near(fitDependence(rows).lambda,truth);
}
for(const v of [NaN,Infinity,-.1,1.1])assert.throws(()=>binaryJoint(.5,.2,.8,v));
assert.throws(()=>fitDependence([]));assert.throws(()=>fitDependence([{base:.5,delta:1,y:1,weight:1}]));
console.log(`PASS: ${cases} joint/marginal/relabeling/gain-scaling checks; five identifiable optima, ownership, oracle traps and invalid inputs`);
