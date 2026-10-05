import assert from 'node:assert/strict';
import {noiseEnvelope,multiplyAffine} from './noise-envelope.mjs';
const types=[0,1,2,7],noises=[.1,.11,.15,.2,.25,.29,.3];
function basis(n,k,u){let choose=1;for(let i=1;i<=k;i++)choose*= (n-i+1)/i;return choose*u**k*(1-u)**(n-k);}
function value(c,u){return c.reduce((s,x,k)=>s+x*basis(c.length-1,k,u),0);}
let checks=0,maxError=0;
for(const counts of [[1,1,1,3],[2,2,1,1],[3,1,1,1]]){
  const schedule=types.map(t=>[0,t,0]);types.forEach((t,i)=>{for(let j=0;j<counts[i];j++)schedule.push([1,t,j]);});
  for(let bits=0;bits<1024;bits+=17){
    const history=schedule.map((action,i)=>({action,outcome:!!(bits&(1<<(9-i)))}));
    const models=noiseEnvelope(history);
    for(let mask=0;mask<16;mask++)for(const noise of noises){
      const weights=Array.from({length:16},(_,h)=>{
        let w=1/16;const roots={};
        for(const {action:[kind,t],outcome} of history){
          if(kind===1&&(mask&(1<<types.indexOf(t)))){w*=Number(outcome===roots[t]);continue;}
          const bit=t===7?((h%2)^Math.floor(h/4)%2):Math.floor(h/2**t)%2;
          const p=bit?1-(t===2?.01:noise):(t===2?.01:noise);
          w*=outcome?p:1-p;if(kind===0)roots[t]=outcome;
        }
        return w;
      });
      const mass=weights.reduce((s,w)=>s+w,0),model=models.find(m=>m.mask===mask);
      if(!model){assert.equal(mass,0);continue;}
      assert(mass>0);const u=(noise-.1)/.2;
      weights.forEach((w,h)=>{const error=Math.abs(w-value(model.coefficients[h],u));assert(error<1e-14);maxError=Math.max(maxError,error);checks++;});
      const denominator=value(model.total,u);
      for(let y=0;y<4;y++){
        const direct=weights.reduce((s,w,h)=>s+(h%4===y?w:0),0)/mass;
        const mixture=model.laws.reduce((s,q,k)=>s+model.total[k]*basis(model.total.length-1,k,u)/denominator*q[y],0);
        const error=Math.abs(direct-mixture);assert(error<1e-12);maxError=Math.max(maxError,error);checks++;
      }
    }
  }
}
assert.deepEqual(multiplyAffine([1],.2,.8),[.2,.8]);
const ordinary=types.map(t=>({action:[0,t,0],outcome:false}));
assert.throws(()=>noiseEnvelope(ordinary.slice(1)));
assert.throws(()=>noiseEnvelope([...ordinary,{action:[1,0,1],outcome:true}]));
assert.throws(()=>noiseEnvelope([...ordinary,{action:[1,0,0],outcome:1}]));
assert.throws(()=>noiseEnvelope(ordinary,types,.3,.1));
assert.throws(()=>noiseEnvelope(ordinary,[0,1,2,3]));
console.log(JSON.stringify({checks,maxError,negativeTests:5}));

