import assert from 'node:assert/strict';

// Binary Brier excess is affine in the unknown outcome probability. Bounding
// both endpoints therefore bounds conditional expected excess without labels.
export function pointwiseGuard(b,c,proposed,epsilon=.01){
  assert([b,c,proposed].every(x=>Number.isFinite(x)&&x>=0&&x<=1));
  assert(Number.isFinite(epsilon)&&epsilon>=0);
  const d=c-b,a=d*d,k=Math.max(2*d*b,2*d*(b-1));
  let w=proposed;
  if(a*w*w+k*w>epsilon)w=epsilon===0?0:Math.min(w,2*epsilon/(k+Math.sqrt(k*k+4*a*epsilon)));
  const p=b+w*d,worst=Math.max((p-b)*(p+b),(p-b)*(p+b-2));
  assert(worst<=epsilon+1e-12);
  return{p,w,proposed,worst,clipped:w<proposed};
}

export function testPointwiseGuard(){
  let checks=0;
  for(const epsilon of [0,.01,1])for(let bi=0;bi<=10;bi++)for(let ci=0;ci<=10;ci++)for(const v of [0,.1,.5,1]){
    const b=bi/10,c=ci/10,f=pointwiseGuard(b,c,v,epsilon);
    assert(f.w<=v&&f.w>=0&&f.p>=0&&f.p<=1);
    for(let qi=0;qi<=10;qi++){const q=qi/10,delta=(f.p-b)*(f.p+b-2*q);assert(delta<=epsilon+1e-12);checks++;}
    if(epsilon===1)assert.equal(f.w,v);
    if(b===c)assert.equal(f.p,b);
  }
  for(const [b,c]of [[.5,.5+1e-14],[0,1e-14],[1,1-1e-14]])assert(Number.isFinite(pointwiseGuard(b,c,1).p));
  for(const args of [[NaN,.5,.5],[.5,Infinity,.5],[.5,.5,-1],[.5,.5,.5,-1]])assert.throws(()=>pointwiseGuard(...args));
  console.log(`pointwise guard endpoint/conditional-risk checks PASS: ${checks}`);
}
