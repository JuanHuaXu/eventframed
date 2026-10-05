import assert from 'node:assert/strict';
import {amplifiedJoint} from './query-amplification.mjs';
import {fitDependence} from './query-dependence.mjs';
const near=(a,b)=>assert(Math.abs(a-b)<1e-10);let cases=0;
for(const p of [1e-6,.01,.2,.5,.9,.999999])for(const f0 of [1e-6,.02,.3,.8,.999999])for(const f1 of [1e-6,.01,.4,.97,.999999])for(const alpha of [0,.2,.7,1]){
  const r=amplifiedJoint(p,f0,f1,alpha),[a,b,c,d]=r.joint;assert(r.joint.every(v=>v>0));near(a+b+c+d,1);near(c+d,p);near(b+d,r.base);assert(r.cap>=1&&r.cap<=2);
  if(alpha===0)assert.deepEqual(r.conditional,[f0,f1]);
  const queryFlip=amplifiedJoint(1-p,f1,f0,alpha),targetFlip=amplifiedJoint(p,1-f0,1-f1,alpha);
  r.joint.forEach((v,i)=>{near(v,queryFlip.joint[(i+2)%4]);near(v,targetFlip.joint[i^1]);});
  near(p*(1-p)*(r.conditional[1]-r.conditional[0])**2,(1+alpha*(r.cap-1))**2*p*(1-p)*(f1-f0)**2);cases++;
}
assert.deepEqual(amplifiedJoint(.3,.5,.5,1).conditional,[.5,.5]);
for(const truth of [0,.25,.5,.75,1]){
  const p=.4,f=[.35,.65],upper=amplifiedJoint(p,...f,1),actual=amplifiedJoint(p,...f,truth),rows=[];
  for(let query=0;query<2;query++)for(let y=0;y<2;y++)rows.push({base:f[query],delta:upper.conditional[query]-f[query],y,weight:(query?p:1-p)*(y?actual.conditional[query]:1-actual.conditional[query])});
  near(fitDependence(rows).lambda,truth);
}
for(const v of [NaN,Infinity,-.1,1.1])assert.throws(()=>amplifiedJoint(.5,.2,.8,v));
console.log(`PASS: ${cases} strict-validity/marginal/relabeling/gain cases; exact original endpoint, zero motion, five identifiable convex optima and invalid inputs`);
