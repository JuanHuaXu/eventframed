import assert from 'node:assert/strict';
export function amplifiedJoint(p,f0,f1,alpha){
  assert([p,f0,f1].every(x=>Number.isFinite(x)&&x>0&&x<1));
  assert(Number.isFinite(alpha)&&alpha>=0&&alpha<=1);
  const base=(1-p)*f0+p*f1,d=[f0-base,f1-base];
  const limit=Math.min(...d.map(x=>x>0?(1-base)/x:x<0?base/-x:Infinity));
  assert(limit>=1-1e-12);
  const cap=Number.isFinite(limit)?Math.min(2,1+.99*(Math.max(1,limit)-1)):2;
  const conditional=[f0,f1].map((f,y)=>f+alpha*(cap-1)*d[y]);
  assert(conditional.every(x=>x>0&&x<1));
  return {base,cap,limit,conditional,joint:[(1-p)*(1-conditional[0]),(1-p)*conditional[0],p*(1-conditional[1]),p*conditional[1]]};
}
