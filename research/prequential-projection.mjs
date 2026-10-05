import assert from 'node:assert/strict';

// Replay adapter only. Future delivery metadata stops at this boundary.
export function projectPrequential(source){
  assert(Array.isArray(source.Steps)&&source.Steps.length===256);
  const issued=[],observed=[];
  for(let j=128;j<160;j++){
    const s=source.Steps[j],p=s.P[10];
    assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512);assert(Number.isFinite(p)&&p>0&&p<1);
    issued.push({origin:j,x:s.X,p});assert.equal(typeof s.Missing,'boolean');
    if(!s.Missing){assert(Number.isInteger(s.Delay)&&s.Delay>=0);const arrived=j+s.Delay;
      if(arrived<=160){assert.equal(typeof s.Y,'boolean');observed.push({origin:j,y:s.Y,arrived});}}
  }
  return {clock:160,begin:128,learner:'v120-issued-segment64',issued,observed};
}
function checkView(v){
  assert.equal(v.clock,160);assert.equal(v.begin,128);assert.equal(v.learner,'v120-issued-segment64');assert.equal(v.issued.length,32);
  v.issued.forEach((r,i)=>{assert.equal(r.origin,128+i);assert(Number.isInteger(r.x)&&r.x>=0&&r.x<512);assert(Number.isFinite(r.p)&&r.p>0&&r.p<1);});
  assert(v.observed.length<=32);let last=127;
  for(const r of v.observed){assert(Number.isInteger(r.origin)&&r.origin>last&&r.origin<160);last=r.origin;assert.equal(typeof r.y,'boolean');assert(Number.isInteger(r.arrived)&&r.arrived>=r.origin&&r.arrived<=160);}
}
const mean=(xs,fallback)=>xs.length?xs.reduce((s,x)=>s+x,0)/xs.length:fallback;
export function prequentialFeatures(view,origin){
  checkView(view);assert(Number.isInteger(origin)&&origin>=152&&origin<160);
  const x=view.issued[origin-128].x;
  const rows=view.observed.map(r=>{const f=view.issued[r.origin-128],res=Number(r.y)-f.p;let bits=x^f.x,distance=0;while(bits){distance+=bits&1;bits>>>=1;}return {origin:r.origin,res,loss:res*res,y:Number(r.y),w:1/(1+distance)};});
  const recent=rows.filter(r=>r.origin>=152),weight=rows.reduce((s,r)=>s+r.w,0);
  const f=[rows.length/32,mean(rows.map(r=>r.y),.5),mean(rows.map(r=>r.loss),.25),.5+.5*mean(rows.map(r=>r.res),0),
    recent.length/8,mean(recent.map(r=>r.loss),.25),weight/32,
    weight?rows.reduce((s,r)=>s+r.w*r.loss,0)/weight:.25,
    .5+.5*(weight?rows.reduce((s,r)=>s+r.w*r.res,0)/weight:0),mean(rows.map(r=>(160-r.origin)/32),1)];
  assert(f.length===10&&f.every(x=>Number.isFinite(x)&&x>=0&&x<=1));return f;
}
