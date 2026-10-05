import assert from 'node:assert/strict';

const floor=1e-12;
const clamp=p=>Math.max(floor,Math.min(1-floor,p));
const sigmoid=z=>z>=0?1/(1+Math.exp(-z)):Math.exp(z)/(1+Math.exp(z));
const grid=[-2,-1,0,1,2];
const z=grid.reduce((s,x)=>s+Math.exp(-x*x/2),0);
const atoms=grid.flatMap(intercept=>grid.map(slopeDelta=>({intercept,slopeDelta,
  prior:.5*Math.exp(-(intercept*intercept+slopeDelta*slopeDelta)/2)/(z*z)
    +(intercept===0&&slopeDelta===0?.5:0)})));
const logit=p=>{assert(Number.isFinite(p)&&p>=0&&p<=1);p=clamp(p);return Math.log(p)-Math.log1p(-p);};
const linear=(a,u)=>a.intercept+(1+a.slopeDelta)*u;

// Exact finite-model posterior, conditional on original issued forecast
// covariates. The identity atom and Gaussian-shaped slab are one frozen prior.
// Averaging the prior is NOT identity calibration, even with no observations.
export function fitFiniteCalibration(rows){
  assert(Array.isArray(rows)&&rows.length<=64);
  const prepared=rows.map(r=>{assert(typeof r.y==='boolean');return {u:logit(r.p),y:r.y};});
  const logs=atoms.map(a=>Math.log(a.prior)+prepared.reduce((s,r)=>{
    const x=r.y?-linear(a,r.u):linear(a,r.u);
    return s-Math.max(x,0)-Math.log1p(Math.exp(-Math.abs(x)));
  },0));
  const max=Math.max(...logs),denom=logs.reduce((s,x)=>s+Math.exp(x-max),0);
  const weights=logs.map(x=>Math.exp(x-max)/denom);
  const predict=p=>{const u=logit(p);return clamp(atoms.reduce((s,a,i)=>s+weights[i]*sigmoid(linear(a,u)),0));};
  return {weights:weights.slice(),logEvidence:max+Math.log(denom),predict};
}
