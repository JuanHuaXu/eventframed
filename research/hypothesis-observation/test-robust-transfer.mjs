import assert from 'node:assert/strict';
import {guardedTransfer} from './robust-transfer.mjs';
let seed=927461;const random=()=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return (seed+.5)/4294967296;};
const simplex=()=>{const p=Array.from({length:4},random),s=p.reduce((a,b)=>a+b,0);return p.map(x=>x/s);};
const risk=(p,q)=>p.reduce((s,x,i)=>s+x*x-2*x*q[i]+q[i],0);
let maxError=0,checks=0;
for(let i=0;i<1000;i++){
  const p=simplex(),r=simplex(),qs=Array.from({length:8},simplex);
  for(const epsilon of [0,.01,.1]){
    const actual=guardedTransfer(p,r,qs,epsilon);
    const regret=l=>Math.max(...qs.map(q=>risk(p.map((x,j)=>x+l*(r[j]-x)),q)-risk(p,q)));
    let lo=0,hi=1;
    if(regret(1)<=epsilon)lo=1;
    else for(let j=0;j<80;j++){const mid=(lo+hi)/2;if(regret(mid)<=epsilon)lo=mid;else hi=mid;}
    maxError=Math.max(maxError,Math.abs(actual.lambda-lo));assert(Math.abs(actual.lambda-lo)<1e-10);
    assert(regret(actual.lambda)<=epsilon+1e-12);checks++;
  }
}
assert.equal(guardedTransfer([.5,.5],[.5,.5],[[1,0]],0).lambda,1);
assert.equal(guardedTransfer([.5,.5],[.9,.1],[[.9,.1]],0).lambda,1);
assert.equal(guardedTransfer([.5,.5],[.9,.1],[[.5,.5]],0).lambda,0);
assert.throws(()=>guardedTransfer([.5,.5],[1,0],[],.01));
assert.throws(()=>guardedTransfer([.5,.5],[1,1],[[1,0]],.01));
console.log(JSON.stringify({checks,maxError,degenerateCases:true,validation:true}));
