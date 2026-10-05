import assert from 'node:assert/strict';
import {projectedTransfer} from './projected-transfer.mjs';
const base=[1,0,0],proposal=[0,1,0],laws=[[1,0,0],[0,1,0],[0,0,1]],t=Math.sqrt(.005),expected=[1-t,t,0];
let maxError=0;
for(const qs of [laws,laws.slice().reverse(),[...laws,...laws]]){
  const r=projectedTransfer(base,proposal,qs);
  r.forecast.forEach((p,i)=>{maxError=Math.max(maxError,Math.abs(p-expected[i]));assert(Math.abs(p-expected[i])<2e-6);});
}
console.log(JSON.stringify({boundaryAndRedundantOrderChecks:3,maxError}));
