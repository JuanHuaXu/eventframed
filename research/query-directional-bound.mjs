import assert from 'node:assert/strict';
const probability=x=>assert(Number.isFinite(x)&&x>=0&&x<=1);
const interval=x=>{assert(Array.isArray(x)&&x.length===2);x.forEach(probability);assert(x[0]<=x[1]);};
// The caller must supply externally justified simultaneous intervals. This
// computes a conditional algebraic bound; it does not certify those intervals.
export function directionalGainBound(base,forecasts,queryInterval,targetIntervals){
  probability(base);assert(Array.isArray(forecasts)&&forecasts.length===2);forecasts.forEach(probability);
  interval(queryInterval);assert(Array.isArray(targetIntervals)&&targetIntervals.length===2);targetIntervals.forEach(interval);
  let lower=Infinity,upper=-Infinity,witness;
  for(const p of queryInterval)for(const q0 of targetIntervals[0])for(const q1 of targetIntervals[1]){
    const delta=[q0,q1].map((q,y)=>(base-forecasts[y])*(base+forecasts[y]-2*q));
    const gain=(1-p)*delta[0]+p*delta[1];
    if(gain<lower){lower=gain;witness={p,q0,q1};}upper=Math.max(upper,gain);
  }
  return {lower,upper,witness};
}
