import assert from 'node:assert/strict';

function probability(p,n){assert(Array.isArray(p)&&p.length===n&&p.every(x=>Number.isFinite(x)&&x>=0&&x<=1));assert(Math.abs(p.reduce((s,x)=>s+x,0)-1)<1e-12);}
// Largest move along the proposed correction whose conditional Brier regret
// is bounded under every declared outcome law. This is model-conditional,
// not an empirical certificate that the ambiguity family includes reality.
export function guardedTransfer(base,proposal,laws,epsilon=.01){
  assert(Array.isArray(base)&&base.length>=2);const n=base.length;
  probability(base,n);probability(proposal,n);
  assert(Array.isArray(laws)&&laws.length>0&&laws.length<=256);
  laws.forEach(p=>probability(p,n));assert(Number.isFinite(epsilon)&&epsilon>=0);
  const d=base.map((p,i)=>proposal[i]-p),a=d.reduce((s,x)=>s+x*x,0);
  const b=Math.max(...laws.map(q=>2*d.reduce((s,x,i)=>s+x*(base[i]-q[i]),0)));
  let lambda=1;
  if(a>0&&a+b>epsilon){
    const disc=Math.sqrt(b*b+4*a*epsilon);
    lambda=b>=0?(epsilon===0?0:2*epsilon/(b+disc)):(disc-b)/(2*a);
    lambda=Math.max(0,Math.min(1,lambda))*(1-1e-14);
  }
  const forecast=base.map((p,i)=>p+lambda*d[i]);probability(forecast,n);
  const worstRegret=a*lambda*lambda+b*lambda;
  assert(worstRegret<=epsilon+1e-12);
  return {forecast,lambda,worstRegret};
}
