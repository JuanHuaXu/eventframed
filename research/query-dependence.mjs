import assert from 'node:assert/strict';
const prob=p=>assert(Number.isFinite(p)&&p>0&&p<1);
export function binaryJoint(p,f0,f1,lambda){
  [p,f0,f1].forEach(prob);assert(Number.isFinite(lambda)&&lambda>=0&&lambda<=1);
  const base=(1-p)*f0+p*f1,a=base+lambda*(f0-base),b=base+lambda*(f1-base);
  return {base,conditional:[a,b],joint:[(1-p)*(1-a),(1-p)*a,p*(1-b),p*b]};
}
export function fitDependence(rows){
  assert(Array.isArray(rows)&&rows.length>0);
  for(const r of rows){prob(r.base);prob(r.base+r.delta);assert(r.y===0||r.y===1);assert(Number.isFinite(r.weight)&&r.weight>0);}
  const derivative=lambda=>rows.reduce((sum,r)=>{
    const p=r.base+lambda*r.delta;
    return sum+r.weight*r.delta*(p-r.y)/(p*(1-p));
  },0);
  const d0=derivative(0),d1=derivative(1);
  if(d0>=0)return {lambda:0,d0,d1};
  if(d1<=0)return {lambda:1,d0,d1};
  let lo=0,hi=1;for(let i=0;i<80;i++){const mid=(lo+hi)/2;if(derivative(mid)>0)hi=mid;else lo=mid;}
  return {lambda:(lo+hi)/2,d0,d1};
}
