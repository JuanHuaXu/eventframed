import assert from 'node:assert/strict';
import {binomialUpper} from './paired-risk-v73.mjs';
// Independently evaluate the Binomial CDF via recursive probability masses.
function cdf(k,n,p){if(p===0)return 1;if(p===1)return Number(k===n);let term=(1-p)**n,sum=term;for(let i=1;i<=k;i++){term*=((n-i+1)/i)*p/(1-p);sum+=term;}return sum;}
const alpha=.05/12;let checks=0;
for(const n of [64,128,256,512])for(const k of [0,1,2,3,5,10,20]){
  const u=binomialUpper(k,n,alpha);assert(Math.abs(cdf(k,n,u)-alpha)<1e-11);assert(cdf(k,n,Math.max(0,u-1e-5))>alpha);assert(cdf(k,n,Math.min(1,u+1e-5))<alpha);checks++;
}
assert(binomialUpper(0,64,alpha)>.02);assert(binomialUpper(0,512,alpha)<.02);
// Exhaustively sum noncoverage probability over every count in small samples.
let coverageChecks=0;
for(const n of [1,4,8,16,32])for(const p of [.001,.01,.02,.1,.3,.5,.9,.99]){
  let failure=0;for(let k=0;k<=n;k++)if(binomialUpper(k,n,alpha)<p){failure+=cdf(k,n,p)-(k?cdf(k-1,n,p):0);}
  assert(failure<=alpha+1e-12);coverageChecks++;
}
console.log(JSON.stringify({status:'PASS',cdfChecks:checks,coverageChecks,alpha,zero64:binomialUpper(0,64,alpha),zero512:binomialUpper(0,512,alpha)}));
