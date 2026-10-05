import assert from 'node:assert/strict';
import {fitResidualLogit,residualLogitDerivatives,solvePositive,probabilityFloor} from './residual-logit.mjs';

const constant=p=>Array(8).fill(p);
for(const full of [false,true]){
  const empty=fitResidualLogit([],full);assert(empty.beta.every(x=>x===0));
  for(const p of [0,1e-12,.01,.5,.99,1-1e-12,1])assert(Math.abs(empty.predict(constant(p))-Math.max(probabilityFloor,Math.min(1-probabilityFloor,p)))<1e-14);
}
// Identical .5 forecasts leave only the intercept. Solve its monotone score
// equation independently by bisection, including all-zero/all-one outcomes.
let roots=0,maxRootError=0;
for(const n of [1,16,64,256])for(const successes of [0,Math.floor(n/3),n])for(const full of [false,true]){
  const rows=Array.from({length:n},(_,i)=>({ps:constant(.5),y:i<successes})),fit=fitResidualLogit(rows,full);
  let lo=-n,hi=n;for(let j=0;j<100;j++){const mid=(lo+hi)/2,g=mid+n/(1+Math.exp(-mid))-successes;if(g>0)hi=mid;else lo=mid;}
  const error=Math.abs(fit.beta[0]-(lo+hi)/2);assert(error<1e-8);assert(fit.beta.slice(1).every(v=>v===0));maxRootError=Math.max(maxRootError,error);roots++;
}
const rows=Array.from({length:13},(_,j)=>({ps:Array.from({length:8},(_,k)=>.03+.94*((j*7+k*3)%23)/22),y:j%3===0}));
let gradientChecks=0,hessianChecks=0,maxGradientError=0,maxHessianError=0;
for(const full of [false,true]){
  const d=full?9:2,beta=Array.from({length:d},(_,j)=>(j-3)*.02),base=residualLogitDerivatives(rows,beta,full),eps=1e-5;
  for(let j=0;j<d;j++){
    const plus=beta.slice(),minus=beta.slice();plus[j]+=eps;minus[j]-=eps;
    const a=residualLogitDerivatives(rows,plus,full),b=residualLogitDerivatives(rows,minus,full);
    const ge=Math.abs((a.objective-b.objective)/(2*eps)-base.g[j]);assert(ge<1e-7);maxGradientError=Math.max(maxGradientError,ge);gradientChecks++;
    for(let i=0;i<d;i++){const he=Math.abs((a.g[i]-b.g[i])/(2*eps)-base.h[i][j]);assert(he<1e-6);maxHessianError=Math.max(maxHessianError,he);hessianChecks++;}
  }
  const solution=solvePositive(base.h,base.g);base.h.forEach((row,i)=>assert(Math.abs(row.reduce((s,x,j)=>s+x*solution[j],0)-base.g[i])<1e-10));
}
let stressFits=0,maxIterations=0;
for(const full of [false,true])for(const y of [false,true])for(const p of [0,1e-12,.01,.99,1-1e-12,1]){
  const rows=Array.from({length:64},()=>({ps:constant(p),y})),fit=fitResidualLogit(rows,full);assert(fit.gradientInf<=1e-8);assert(Number.isFinite(fit.objective));maxIterations=Math.max(maxIterations,fit.iterations);stressFits++;
}
const fit=fitResidualLogit(rows),before=fit.predict(constant(.3));fit.beta[0]=999;assert.equal(fit.predict(constant(.3)),before);
assert.throws(()=>fitResidualLogit([{ps:constant(NaN),y:true}]));assert.throws(()=>fitResidualLogit([{ps:constant(.5),y:1}]));assert.throws(()=>fitResidualLogit(Array(257).fill({ps:constant(.5),y:true})));
console.log(JSON.stringify({roots,maxRootError,gradientChecks,maxGradientError,hessianChecks,maxHessianError,stressFits,maxIterations,ownership:true}));
