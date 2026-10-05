import assert from 'node:assert/strict';
import {predictiveQueryValue} from './predictive-query-value.mjs';
let checks=0;
for(let i=1;i<20;i++)for(let j=1;j<20;j++){
 const a=i/20,b=j/20,py=[.3,.7],conditional=[[a,1-a],[b,1-b]],base=[.3*a+.7*b,1-(.3*a+.7*b)],probes=[[.1,.9],[.8,.2]];
 let before=0,after=0;
 for(const p of probes)for(let y=0;y<2;y++)for(let k=0;k<2;k++)for(const outcome of [0,1]){
  const mass=py[y]*conditional[y][k]*(outcome?p[k]:1-p[k])/2;
  const pb=base.reduce((v,w,k)=>v+w*p[k],0),pa=conditional[y].reduce((v,w,k)=>v+w*p[k],0);
  before+=mass*(outcome-pb)**2;after+=mass*(outcome-pa)**2;
 }
 assert(Math.abs(before-after-predictiveQueryValue(base,{probabilities:py,conditional},probes))<1e-12);checks++;
}
assert.equal(predictiveQueryValue([.5,.5],{probabilities:[.5,.5],conditional:[[1,0],[0,1]]},[[.2,.2]]),0);
assert.throws(()=>predictiveQueryValue([1],{probabilities:[1,0],conditional:[[1],[1]]},[]));
console.log(JSON.stringify({checks,status:'PASS'}));
