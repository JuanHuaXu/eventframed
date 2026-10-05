import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';

// Oracle diagnostic only. True channel laws never enter a runtime learner.
const cases={
  sparse:[[.85,.05],[.85,.05],[0,0],[0,0]],
  homogeneous:Array.from({length:4},()=>[.5,.1]),
  negative:[[.05,.85],[.05,.85],[0,0],[0,0]],
  weak:[[.8,0],[0,0],[0,0],[0,0]],
  boundary:[[1,0],[0,2/15],[0,2/15],[0,2/15]],
  zero:Array.from({length:4},()=>[0,0]),
};
const rows=[];
for(const [name,law]of Object.entries(cases)){
  const m=law.map(([p,n])=>p-n),variances=law.map(([p,n],i)=>p+n-m[i]**2);
  const total=variances.reduce((s,v)=>s+Math.sqrt(Math.max(0,v)),0);
  const q=variances.map(v=>total===0?.25:.1+.6*Math.sqrt(Math.max(0,v))/total);
  const average=m.reduce((a,b)=>a+b,0)/4,w=q.map(v=>.25/v),c=m.map((v,i)=>average-w[i]*v);
  const eta=Math.min(1,...c.map((v,i)=>v===0?Infinity:(3.53-w[i])/Math.abs(v)));
  assert.ok(Math.abs(q.reduce((a,b)=>a+b,0)-1)<1e-12);
  assert.ok(Math.abs(q.reduce((s,v,i)=>s+v*eta*c[i],0))<1e-12);
  for(const sign of [1,-1]){
    const endpoints=w.flatMap((v,i)=>[-1,1].map(d=>sign*(v*d+eta*c[i])-.15));
    const lower=Math.min(...endpoints);
    const cap=Math.min(.8,lower<0?.92/(-lower):.8);
    const tape=law.flatMap(([p,n],i)=>[[-1,n],[0,1-p-n],[1,p]].map(([d,prob])=>({p:q[i]*prob,x:sign*(w[i]*d+eta*c[i])-.15})));
    assert.ok(Math.abs(tape.reduce((s,t)=>s+t.p,0)-1)<1e-12);
    const growth=l=>tape.reduce((s,t)=>s+t.p*Math.log1p(l*t.x),0);
    const derivative=l=>tape.reduce((s,t)=>s+t.p*t.x/(1+l*t.x),0);
    let rate;
    if(derivative(0)<=1e-12)rate=0;
    else if(derivative(cap)>=0)rate=cap;
    else{let lo=0,hi=cap;for(let i=0;i<80;i++){const mid=(lo+hi)/2;if(derivative(mid)>0)lo=mid;else hi=mid;}rate=(lo+hi)/2;}
    assert.ok(cap>=.25-1e-12);
    for(const x of endpoints)assert.ok(1+rate*x>=.08-1e-12);
    for(let i=0;i<=10000;i++)assert.ok(growth(rate)>=growth(cap*i/10000)-1e-12,'optimizer below grid maximum');
    rows.push({name,sign,q,m,eta,cap,rate,fixedGrowth:growth(.25),oracleGrowth:growth(rate),minimumContractFactor:Math.min(...endpoints.map(x=>1+rate*x))});
  }
}
const sources={};for(const p of ['research/bet-headroom-v75.mjs','docs/experiments/mmm-bet-headroom-v75-protocol.md'])sources[p]=crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const result={scope:'deterministic known-law growth diagnostic, not stopping-time or runtime performance',sources,rows};
if(process.argv[2])fs.writeFileSync(process.argv[2],JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(result,null,2));
